package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"social-network/internal/models"

	"github.com/lib/pq"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrInvalidPage  = errors.New("invalid page number")
)

const (
	// Базовые запросы для выборки постов
	basePostSelect = `
        SELECT p.id, p.author_id, p.content, p.created_at, p.updated_at,
               u.username, u.role,
               ARRAY_AGG(h.name) FILTER (WHERE h.name IS NOT NULL) as hashtags
        FROM posts p
        JOIN users u ON p.author_id = u.id
        LEFT JOIN post_hashtags ph ON p.id = ph.post_id
        LEFT JOIN hashtags h ON ph.hashtag_id = h.id`

	baseUserPostsQuery = basePostSelect + `
        WHERE p.author_id = $1
        GROUP BY p.id, u.id`

	baseUserPostsQueryHashtag = `
		SELECT p.id, p.author_id, p.content, p.created_at, p.updated_at,
			u.username, u.role,
			ARRAY_AGG(h.name) FILTER (WHERE h.name IS NOT NULL) as hashtags
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN post_hashtags ph ON p.id = ph.post_id
		LEFT JOIN hashtags h ON ph.hashtag_id = h.id
		WHERE p.author_id = $1
		GROUP BY p.id, u.id
		HAVING $4::varchar IS NULL OR $4 = ANY(ARRAY_AGG(h.name) FILTER (WHERE h.name IS NOT NULL))`

	baseFeedPostsQuery = basePostSelect + `
        JOIN followers f ON p.author_id = f.following_id
        WHERE f.follower_id = $1
        GROUP BY p.id, u.id`

	baseFeedPostsQueryHashtag = `
		SELECT p.id, p.author_id, p.content, p.created_at, p.updated_at,
			u.username, u.role,
			ARRAY_AGG(h.name) FILTER (WHERE h.name IS NOT NULL) as hashtags
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN post_hashtags ph ON p.id = ph.post_id
		LEFT JOIN hashtags h ON ph.hashtag_id = h.id
		JOIN followers f ON p.author_id = f.following_id
		WHERE f.follower_id = $1
		GROUP BY p.id, u.id
		HAVING $4::varchar IS NULL OR $4 = ANY(ARRAY_AGG(h.name) FILTER (WHERE h.name IS NOT NULL))`

	baseHashtagPostsQuery = basePostSelect + `
        WHERE h.name = $1
        GROUP BY p.id, u.id`

	// Запросы для создания и обновления
	createPostQuery = `
        INSERT INTO posts (author_id, content)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at`

	insertHashtagsQuery = `
        INSERT INTO hashtags (name)
        SELECT unnest($1::text[])
        ON CONFLICT (name) DO NOTHING`

	linkHashtagsQuery = `
        INSERT INTO post_hashtags (post_id, hashtag_id)
        SELECT $1, id FROM hashtags 
        WHERE name = ANY($2)`

	updatePostQuery = `
        UPDATE posts 
        SET content = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING updated_at`

	deletePostQuery = `DELETE FROM posts WHERE id = $1`
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаем пост
	err = tx.QueryRow(
		createPostQuery,
		post.AuthorID,
		post.Content,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	// Обрабатываем хэштеги
	if err := r.handleHashtags(tx, post.ID, post.Hashtags); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostRepository) GetByID(postID int) (*models.Post, error) {
	post := &models.Post{
		Author: &models.User{},
	}

	query := basePostSelect + ` WHERE p.id = $1 GROUP BY p.id, u.id`
	err := r.db.QueryRow(query, postID).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Author.Username,
		&post.Author.Role,
		pq.Array(&post.Hashtags),
	)

	if err == sql.ErrNoRows {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query post: %w", err)
	}

	return post, nil
}

func (r *PostRepository) GetUserPosts(userID int, page, perPage int, orderDesc bool, hashtag *string) ([]models.Post, error) {
	if err := validatePagination(page, perPage); err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage

	var rows *sql.Rows
	var err error
	if hashtag == nil {
		query := buildPaginatedQuery(baseUserPostsQuery, orderDesc)
		rows, err = r.db.Query(query, userID, perPage, offset)
	} else {
		query := buildPaginatedQuery(baseUserPostsQueryHashtag, orderDesc)
		rows, err = r.db.Query(query, userID, perPage, offset, hashtag)
	}

	if err != nil {
		return nil, fmt.Errorf("query posts: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

func (r *PostRepository) GetFeedPosts(userID int, page, perPage int, orderDesc bool, hashtag *string) ([]models.Post, error) {
	if err := validatePagination(page, perPage); err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage

	var rows *sql.Rows
	var err error
	if hashtag == nil {
		query := buildPaginatedQuery(baseFeedPostsQuery, orderDesc)
		rows, err = r.db.Query(query, userID, perPage, offset)
	} else {
		query := buildPaginatedQuery(baseFeedPostsQueryHashtag, orderDesc)
		rows, err = r.db.Query(query, userID, perPage, offset, hashtag)
	}

	if err != nil {
		return nil, fmt.Errorf("query feed: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

func (r *PostRepository) Update(post *models.Post) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Обновляем пост
	err = tx.QueryRow(updatePostQuery, post.Content, post.ID).Scan(&post.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrPostNotFound
	}
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}

	// Удаляем старые хэштеги
	_, err = tx.Exec("DELETE FROM post_hashtags WHERE post_id = $1", post.ID)
	if err != nil {
		return fmt.Errorf("delete old hashtags: %w", err)
	}

	// Добавляем новые хэштеги
	if err := r.handleHashtags(tx, post.ID, post.Hashtags); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostRepository) Delete(postID int) error {
	result, err := r.db.Exec(deletePostQuery, postID)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return ErrPostNotFound
	}

	return nil
}

func (r *PostRepository) handleHashtags(tx *sql.Tx, postID int, hashtags []string) error {
	if len(hashtags) == 0 {
		return nil
	}

	// Вставляем хэштеги
	_, err := tx.Exec(insertHashtagsQuery, pq.Array(hashtags))
	if err != nil {
		return fmt.Errorf("insert hashtags: %w", err)
	}

	// Связываем с постом
	_, err = tx.Exec(linkHashtagsQuery, postID, pq.Array(hashtags))
	if err != nil {
		return fmt.Errorf("link hashtags: %w", err)
	}

	return nil
}

func (r *PostRepository) scanPosts(rows *sql.Rows) ([]models.Post, error) {
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		post.Author = &models.User{}

		err := rows.Scan(
			&post.ID,
			&post.AuthorID,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Author.Username,
			&post.Author.Role,
			pq.Array(&post.Hashtags),
		)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return posts, nil
}

func (r *PostRepository) GetPostsByHashtag(hashtag string, searchType string, username string, page, perPage int, orderDesc bool) ([]models.Post, error) {
	if err := validatePagination(page, perPage); err != nil {
		return nil, err
	}

	var baseQuery string
	var args []interface{}
	args = append(args, hashtag)

	switch searchType {
	case models.SearchTypeUser:
		baseQuery = basePostSelect + `
            WHERE h.name = $1 AND u.username = $2
            GROUP BY p.id, u.id`
		args = append(args, username)
	case models.SearchTypeFollowing:
		baseQuery = basePostSelect + `
            JOIN followers f ON p.author_id = f.following_id
            WHERE h.name = $1 AND f.follower_id = $2
            GROUP BY p.id, u.id`
		args = append(args, username)
	default: // SearchTypeAll
		baseQuery = baseHashtagPostsQuery
	}

	query := buildPaginatedQuery(baseQuery, orderDesc)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query hashtag posts: %w", err)
	}
	defer rows.Close()

	return r.scanPosts(rows)
}

func buildPaginatedQuery(baseQuery string, orderDesc bool) string {
	orderBy := "ASC"
	if orderDesc {
		orderBy = "DESC"
	}
	return fmt.Sprintf(`
        %s
        ORDER BY p.created_at %s
        LIMIT $2 OFFSET $3
    `, baseQuery, orderBy)
}

func validatePagination(page, perPage int) error {
	if page < 1 || perPage < 1 {
		return ErrInvalidPage
	}
	return nil
}
