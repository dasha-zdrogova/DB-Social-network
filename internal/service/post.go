package service

import (
	"errors"
	"social-network/internal/models"
	"social-network/internal/repository/postgres"
	"strings"
	"time"
	"unicode"
)

var (
	ErrNotYourPost = errors.New("do not have access to post")
)

type PostService interface {
	Create(userID int, content string) (*models.Post, error)
	GetByID(postID int) (*models.Post, error)
	Update(postID, userID int, content string) error
	Delete(postID, userID int) error
	GetUserPosts(userID int, page, perPage int, orderDesc bool, hashtag *string) ([]models.Post, error)
	GetFeed(userID int, page, perPage int, orderDesc bool, hashtag *string) ([]models.Post, error)
	GetPostsByHashtag(hashtag string, page, perPage int, orderDesc bool) ([]models.Post, error)
	GetMyPosts(userID int, page, perPage int, orderDesc bool) ([]models.Post, error)
}

type PostServiceImpl struct {
	repos *postgres.Repositories
}

func NewPostService(repos *postgres.Repositories) PostService {
	return &PostServiceImpl{repos: repos}
}

func (s *PostServiceImpl) Create(userID int, content string) (*models.Post, error) {
	user, err := s.repos.Users.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, postgres.ErrUserNotFound
	}
	// Create post object
	post := &models.Post{
		AuthorID:  userID,
		Author:    user,
		Content:   content,
		Hashtags:  extractHashtags(content),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to repository
	if err := s.repos.Posts.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostServiceImpl) GetByID(postID int) (*models.Post, error) {
	return s.repos.Posts.GetByID(postID)
}

func (s *PostServiceImpl) Update(postID, userID int, content string) error {
	// Get post
	post, err := s.repos.Posts.GetByID(postID)
	if err != nil {
		return err
	}

	// Get user with role using existing method
	user, err := s.repos.Users.GetByID(userID)
	if err != nil {
		return err
	}

	if !(post.AuthorID == userID ||
		user.Role == models.RoleAdmin ||
		user.Role == models.RoleModerator) {
		return ErrNotYourPost
	}

	post.Content = content
	post.Hashtags = extractHashtags(content)
	post.UpdatedAt = time.Now()
	return s.repos.Posts.Update(post)
}

func (s *PostServiceImpl) Delete(postID, userID int) error {
	// Get post
	post, err := s.repos.Posts.GetByID(postID)
	if err != nil {
		return err
	}

	// Get user with role using existing GetByID
	user, err := s.repos.Users.GetByID(userID)
	if err != nil {
		return err
	}

	// Check if user is owner or admin
	if !(post.AuthorID == userID ||
		user.Role == models.RoleAdmin) {
		return ErrNotYourPost
	}

	return s.repos.Posts.Delete(postID)
}

func (s *PostServiceImpl) GetUserPosts(userID int, page, perPage int, orderDesc bool, hashtag *string) ([]models.Post, error) {
	return s.repos.Posts.GetUserPosts(userID, page, perPage, orderDesc, hashtag)
}

func (s *PostServiceImpl) GetFeed(userID int, page, perPage int, orderDesc bool, hashtag *string) ([]models.Post, error) {
	return s.repos.Posts.GetFeedPosts(userID, page, perPage, orderDesc, hashtag)
}

func (s *PostServiceImpl) GetPostsByHashtag(hashtag string, page, perPage int, orderDesc bool) ([]models.Post, error) {
	return s.repos.Posts.GetPostsByHashtag(hashtag, models.SearchTypeAll, "", page, perPage, orderDesc)
}

func (s *PostServiceImpl) GetMyPosts(userID int, page, perPage int, orderDesc bool) ([]models.Post, error) {
	return s.GetUserPosts(userID, page, perPage, orderDesc, nil)
}

func extractHashtags(content string) []string {
	words := strings.Fields(content)

	hashtagMap := make(map[string]struct{})

	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			hashtag := strings.TrimLeft(word, "#")
			hashtag = strings.TrimRight(hashtag, ".,!?")

			if len(hashtag) > 0 && isValidHashtag(hashtag) {
				hashtagMap[hashtag] = struct{}{}
			}
		}
	}

	hashtags := make([]string, 0, len(hashtagMap))
	for hashtag := range hashtagMap {
		hashtags = append(hashtags, hashtag)
	}

	return hashtags
}

func isValidHashtag(hashtag string) bool {
	for _, r := range hashtag {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' {
			return false
		}
	}
	return true
}
