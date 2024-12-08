package service

import (
	"social-network/internal/models"
	"social-network/internal/repository/postgres"
)

type FollowService interface {
	Follow(followerID, followingID int) error
	Unfollow(followerID, followingID int) error
	GetFollowers(userID int, page int) ([]models.User, error)
}

type FollowServiceImpl struct {
	repo *postgres.FollowRepository
}

func NewFollowService(repo *postgres.FollowRepository) FollowService {
	return &FollowServiceImpl{repo: repo}
}

func (s *FollowServiceImpl) Follow(followerID, followingID int) error {
	// TODO: implement
	return nil
}

func (s *FollowServiceImpl) Unfollow(followerID, followingID int) error {
	// TODO: implement
	return nil
}

func (s *FollowServiceImpl) GetFollowers(userID int, page int) ([]models.User, error) {
	// TODO: implement
	return nil, nil
}

func (s *FollowServiceImpl) GetFollowing(userID int, page int) ([]models.User, error) {
	// TODO: implement
	return nil, nil
}

func (s *FollowServiceImpl) GetMutualFollows(userID int, page int) ([]models.User, error) {
	// TODO: implement
	return nil, nil
}
