package service

import "social-network/internal/repository/postgres"

type Services struct {
	Users   UserService
	Posts   PostService
	Follows FollowService
}

func NewServices(repos *postgres.Repositories) *Services {
	return &Services{
		Users:   NewUserService(repos.Users),
		Posts:   NewPostService(repos),
		Follows: NewFollowService(repos.Follows),
	}
}
