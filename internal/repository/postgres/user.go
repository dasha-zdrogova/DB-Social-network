package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"social-network/internal/models"

	"github.com/lib/pq"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
        INSERT INTO users (username, password_hash, role)
        VALUES ($1, crypt($2, gen_salt('bf')), $3)
        RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		user.Username,
		user.PasswordHash,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if isPgDuplicateError(err) {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (r *UserRepository) ValidatePassword(username, password string) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, username, role, created_at, updated_at
        FROM users
        WHERE username = $1 
        AND password_hash = crypt($2, password_hash)`

	err := r.db.QueryRow(query, username, password).Scan(
		&user.ID,
		&user.Username,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, username, password_hash, role, created_at, updated_at
        FROM users
        WHERE username = $1`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, username, password_hash, role, created_at, updated_at
        FROM users
        WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Delete(userID int, withPosts bool) error {
	result, err := r.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) UpdateRole(userID int, role models.UserRole) error {
	query := `
        UPDATE users
        SET role = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2`

	result, err := r.db.Exec(query, role, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) CreateFollow(followerID, followingID int) error {
	query := `
        INSERT INTO followers (follower_id, following_id)
        VALUES ($1, $2)`

	_, err := r.db.Exec(query, followerID, followingID)
	if err != nil {
		if isPgDuplicateError(err) {
			return ErrAlreadyFollowing
		}
		return fmt.Errorf("create follow: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteFollow(followerID, followingID int) error {
	query := `
        DELETE FROM followers 
        WHERE follower_id = $1 AND following_id = $2`

	result, err := r.db.Exec(query, followerID, followingID)
	if err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFollowing
	}
	return nil
}

func (r *UserRepository) GetFollowers(userID, page, perPage int) ([]models.User, error) {
	query := `
        SELECT u.id, u.username, u.role, u.created_at, u.updated_at
        FROM users u
        JOIN followers f ON u.id = f.follower_id
        WHERE f.following_id = $1
        ORDER BY u.username
        LIMIT $2 OFFSET $3`

	offset := (page - 1) * perPage
	rows, err := r.db.Query(query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query followers: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) GetFollowing(userID, page, perPage int) ([]models.User, error) {
	query := `
        SELECT u.id, u.username, u.role, u.created_at, u.updated_at
        FROM users u
        JOIN followers f ON u.id = f.following_id
        WHERE f.follower_id = $1
        ORDER BY u.username
        LIMIT $2 OFFSET $3`

	offset := (page - 1) * perPage
	rows, err := r.db.Query(query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query following: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) GetMutualFollows(userID, page, perPage int) ([]models.User, error) {
	query := `
        SELECT u.id, u.username, u.role, u.created_at, u.updated_at
        FROM users u
        JOIN followers f1 ON u.id = f1.following_id
        JOIN followers f2 ON u.id = f2.follower_id
        WHERE f1.follower_id = $1 AND f2.following_id = $1
        ORDER BY u.username
        LIMIT $2 OFFSET $3`

	offset := (page - 1) * perPage
	rows, err := r.db.Query(query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query mutual follows: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) scanUsers(rows *sql.Rows) ([]models.User, error) {
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
	return users, rows.Err()
}

func isPgDuplicateError(err error) bool {
	pgErr, ok := err.(*pq.Error)
	return ok && pgErr.Code == "23505"
}
