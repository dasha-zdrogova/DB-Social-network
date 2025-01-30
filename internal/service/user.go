package service

import (
	"errors"
	"social-network/internal/models"
	"social-network/internal/repository/postgres"
)

var (
	ErrInvalidCredentials = errors.New("user not found")
	ErrNotAdmin           = errors.New("you are not admin")
	ErrUnknownRole        = errors.New("unknown role")
)

type UserService interface {
	Register(username, password string) error
	ValidateUser(username, password string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	Follow(followerID, followingID int) error
	Unfollow(followerID, followingID int) error
	GetFollowers(userID int, page, perPage int) ([]models.User, error)
	GetFollowing(userID int, page, perPage int) ([]models.User, error)
	GetMutualFollows(userID int, page, perPage int) ([]models.User, error)
	DeleteAccount(userID, toDeleteID int, withPosts bool) error
	UpdateRole(adminID int, targetUserID int, newRole string) error
}

type UserServiceImpl struct {
	repo *postgres.UserRepository
}

func NewUserService(repo *postgres.UserRepository) UserService {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) ValidateUser(username, password string) (*models.User, error) {
	return s.repo.ValidatePassword(username, password)
}

func (s *UserServiceImpl) Register(username, password string) error {
	user := &models.User{
		Username:     username,
		PasswordHash: password,
		Role:         models.RoleUser,
	}
	return s.repo.Create(user)
}

func (s *UserServiceImpl) GetUserByUsername(username string) (*models.User, error) {
	return s.repo.GetByUsername(username)
}

func (s *UserServiceImpl) Follow(followerID, followingID int) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}
	return s.repo.CreateFollow(followerID, followingID)
}

func (s *UserServiceImpl) Unfollow(followerID, followingID int) error {
	return s.repo.DeleteFollow(followerID, followingID)
}

func (s *UserServiceImpl) GetFollowers(userID int, page, perPage int) ([]models.User, error) {
	return s.repo.GetFollowers(userID, page, perPage)
}

func (s *UserServiceImpl) GetFollowing(userID int, page, perPage int) ([]models.User, error) {
	return s.repo.GetFollowing(userID, page, perPage)
}

func (s *UserServiceImpl) GetMutualFollows(userID int, page, perPage int) ([]models.User, error) {
	return s.repo.GetMutualFollows(userID, page, perPage)
}

func (s *UserServiceImpl) DeleteAccount(userID, toDeleteID int, withPosts bool) error {
	if userID == toDeleteID {
		return s.repo.Delete(toDeleteID, withPosts)
	}

	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	if !(user.Role == models.RoleAdmin || user.Role == models.RoleModerator) {
		return ErrNotAdmin
	}

	return s.repo.Delete(toDeleteID, withPosts)
}

func (s *UserServiceImpl) UpdateRole(adminID int, targetUserID int, newRole string) error {
	admin, err := s.repo.GetByID(adminID)
	if err != nil {
		return err
	}
	if admin.Role != models.RoleAdmin {
		return ErrNotAdmin
	}

	newRoleModel := getRole(newRole)
	if newRoleModel == nil {
		return ErrUnknownRole
	}

	return s.repo.UpdateRole(targetUserID, *newRoleModel)
}

func getRole(role string) *models.UserRole {
	validRoles := map[string]models.UserRole{
		"user":      models.RoleUser,
		"moderator": models.RoleModerator,
		"admin":     models.RoleAdmin,
	}
	modelsRole, ok := validRoles[role]
	if !ok {
		return nil
	} else {
		return &modelsRole
	}
}
