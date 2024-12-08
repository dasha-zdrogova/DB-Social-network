package service

import (
	"errors"
	"social-network/internal/models"
	"social-network/internal/repository/postgres"
)

var (
	ErrInvalidCredentials = errors.New("user not found")
)

type UserService interface {
	Register(username, password string) error
	ValidateUser(username, password string) (*models.User, error)
}

type UserServiceImpl struct {
	repo *postgres.UserRepository
}

func NewUserService(repo *postgres.UserRepository) UserService {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) Register(username, password string) error {
	// TODO: implement
	return nil
}

func (s *UserServiceImpl) ValidateUser(username, password string) (*models.User, error) {
	// TODO: implement
	return nil, nil
}
