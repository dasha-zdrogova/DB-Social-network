package repository

import "social-network/internal/models"

type UserRepository interface {
	Create(user *models.User) error
	GetByUsername(username string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	Delete(userID int, withPosts bool) error
	UpdateRole(userID int, role models.UserRole) error
	CreateFollow(followerID, followingID int) error
	DeleteFollow(followerID, followingID int) error
	GetFollowers(userID, page, perPage int) ([]models.User, error)
	GetFollowing(userID, page, perPage int) ([]models.User, error)
	GetMutualFollows(userID, page, perPage int) ([]models.User, error)
}

type PostRepository interface {
	Create(post *models.Post) error
	GetByID(postID int) (*models.Post, error)
	Update(post *models.Post) error
	Delete(postID int) error
	GetUserPosts(userID int, page, perPage int, orderDesc bool) ([]models.Post, error)
	GetFeedPosts(userID int, page, perPage int, orderDesc bool) ([]models.Post, error)
	GetPostsByHashtag(hashtag string, searchType string, username string, page, perPage int, orderDesc bool) ([]models.Post, error)
}

type FollowRepository interface {
	Follow(followerID, followingID int) error
	Unfollow(followerID, followingID int) error
	GetFollowers(userID int, page, perPage int) ([]models.User, error)
	GetFollowing(userID int, page, perPage int) ([]models.User, error)
	GetMutualFollows(userID int, page, perPage int) ([]models.User, error)
}
