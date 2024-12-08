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

// Аутентификация
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Управление аккаунтом
type DeleteAccountRequest struct {
	WithPosts bool `json:"with_posts"`
}

// Посты
type CreatePostRequest struct {
	Content string `json:"content" validate:"required"`
}

type UpdatePostRequest struct {
	PostID  int    `json:"post_id" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type GetMyPostsRequest struct {
	Page      int  `json:"page" default:"1"`
	PerPage   int  `json:"per_page" default:"20"`
	OrderDesc bool `json:"order_desc" default:"true"`
}

type GetMyPostsResponse struct {
	Posts      []Post `json:"posts"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	HasMore    bool   `json:"has_more"`
}

type GetUserPostsRequest struct {
	Username  string `json:"username" validate:"required"`
	Page      int    `json:"page" default:"1"`
	PerPage   int    `json:"per_page" default:"20"`
	OrderDesc bool   `json:"order_desc" default:"true"`
}

type GetUserPostsResponse struct {
	Posts      []Post `json:"posts"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	HasMore    bool   `json:"has_more"`
}

type DeletePostRequest struct {
	PostID int `json:"post_id" validate:"required"`
}

// Подписки
type FollowRequest struct {
	Username string `json:"username" validate:"required"`
}

type FollowersResponse struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	HasMore    bool   `json:"has_more"`
}

type FollowingResponse struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	HasMore    bool   `json:"has_more"`
}

type MutualFollowsResponse struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	HasMore    bool   `json:"has_more"`
}

type GetFeedResponse struct {
	Posts      []Post `json:"posts"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	HasMore    bool   `json:"has_more"`
}

type GetPostsByHashtagRequest struct {
	Hashtag    string `json:"hashtag" validate:"required"`
	SearchType string `json:"type" enum:"user,following,all" default:"all"`
	Username   string `json:"username,omitempty"`
	Page       int    `json:"page" default:"1"`
	PerPage    int    `json:"per_page" default:"20"`
	OrderDesc  bool   `json:"order_desc" default:"true"`
}

type GetPostsByHashtagResponse struct {
	Posts      []Post `json:"posts"`
	Hashtag    string `json:"hashtag"`
	SearchType string `json:"search_type"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	HasMore    bool   `json:"has_more"`
}

// Типы поиска
const (
	SearchTypeUser      = "user"      // посты конкретного пользователя
	SearchTypeFollowing = "following" // посты подписок
	SearchTypeAll       = "all"       // посты всех пользователей
)

// Модерация
type ModeratePostRequest struct {
	PostID  int    `json:"post_id" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type DeleteUserRequest struct {
	Username  string `json:"username" validate:"required"`
	WithPosts bool   `json:"with_posts"`
}

type ModeratorDeletePostRequest struct {
	PostID int `json:"post_id" validate:"required"`
}

// Администрирование
type UpdateUserRoleRequest struct {
	Username string   `json:"username" validate:"required"`
	NewRole  UserRole `json:"new_role" validate:"required"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
