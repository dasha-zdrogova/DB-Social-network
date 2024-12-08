package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"social-network/internal/models"
)

var (
	ErrAlreadyFollowing = errors.New("already following this user")
	ErrNotFollowing     = errors.New("not following this user")
)

const (
	createFollowQuery = `
        INSERT INTO followers (follower_id, following_id)
        VALUES ($1, $2)`

	deleteFollowQuery = `
        DELETE FROM followers 
        WHERE follower_id = $1 AND following_id = $2`

	baseFollowSelect = `
        SELECT u.id, u.username, u.role, u.created_at, u.updated_at
        FROM users u`

	getFollowersQuery = baseFollowSelect + `
        JOIN followers f ON u.id = f.follower_id
        WHERE f.following_id = $1
        ORDER BY u.username
        LIMIT $2 OFFSET $3`

	getFollowingQuery = baseFollowSelect + `
        JOIN followers f ON u.id = f.following_id
        WHERE f.follower_id = $1
        ORDER BY u.username
        LIMIT $2 OFFSET $3`

	getMutualFollowsQuery = baseFollowSelect + `
        JOIN followers f1 ON u.id = f1.following_id
        JOIN followers f2 ON u.id = f2.follower_id
        WHERE f1.follower_id = $1 AND f2.following_id = $1
        ORDER BY u.username
        LIMIT $2 OFFSET $3`
)

type FollowRepository struct {
	db *sql.DB
}

func NewFollowRepository(db *sql.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Follow(followerID, followingID int) error {
	_, err := r.db.Exec(createFollowQuery, followerID, followingID)
	if err != nil {
		if isPgDuplicateError(err) {
			return ErrAlreadyFollowing
		}
		return fmt.Errorf("create follow: %w", err)
	}
	return nil
}

func (r *FollowRepository) Unfollow(followerID, followingID int) error {
	result, err := r.db.Exec(deleteFollowQuery, followerID, followingID)
	if err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFollowing
	}

	return nil
}

func (r *FollowRepository) GetFollowers(userID int, page, perPage int) ([]models.User, error) {
	if err := validatePagination(page, perPage); err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(getFollowersQuery, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query followers: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *FollowRepository) GetFollowing(userID int, page, perPage int) ([]models.User, error) {
	if err := validatePagination(page, perPage); err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(getFollowingQuery, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query following: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *FollowRepository) GetMutualFollows(userID int, page, perPage int) ([]models.User, error) {
	if err := validatePagination(page, perPage); err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(getMutualFollowsQuery, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query mutual follows: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *FollowRepository) scanUsers(rows *sql.Rows) ([]models.User, error) {
	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return users, nil
}
