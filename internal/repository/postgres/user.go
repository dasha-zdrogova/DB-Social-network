package postgres

import (
	"database/sql"
	"errors"
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
        VALUES ($1, $2, $3)
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

func (r *UserRepository) Delete(userID int) error {
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

func isPgDuplicateError(err error) bool {
	pgErr, ok := err.(*pq.Error)
	return ok && pgErr.Code == "23505"
}
