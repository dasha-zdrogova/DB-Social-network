package service

import (
	"social-network/internal/models"
	"social-network/internal/repository/postgres"
)

type PostService interface {
	Create(userID int, content string) (*models.Post, error)
	GetByID(postID int) (*models.Post, error)
	GetUserPosts(userID int, page int) ([]models.Post, error)
	GetFeed(userID int, page int) ([]models.Post, error)
}

type PostServiceImpl struct {
	repo *postgres.PostRepository
}

func NewPostService(repo *postgres.PostRepository) PostService {
	return &PostServiceImpl{repo: repo}
}

func (s *PostServiceImpl) Create(userID int, content string) (*models.Post, error) {
	// TODO: implement
	return nil
}

func (s *PostServiceImpl) GetByID(postID int) (*models.Post, error) {
	// TODO: implement
	return nil, nil
}

func (s *PostServiceImpl) GetUserPosts(userID int, page int) ([]models.Post, error) {
	// TODO: implement
	return nil, nil
}

func (s *PostServiceImpl) GetFeed(userID int, page int) ([]models.Post, error) {
	// TODO: implement
	return nil, nil
}
