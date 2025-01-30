package models

import "time"

type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleModerator UserRole = "moderator"
	RoleAdmin     UserRole = "admin"
)

// Основные сущности
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Post struct {
	ID        int       `json:"id"`
	AuthorID  int       `json:"author_id"`
	Author    *User     `json:"author,omitempty"`
	Content   string    `json:"content"`
	Hashtags  []string  `json:"hashtags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Типы поиска
const (
	SearchTypeUser      = "user"      // посты конкретного пользователя
	SearchTypeFollowing = "following" // посты подписок
	SearchTypeAll       = "all"       // посты всех пользователей
)
