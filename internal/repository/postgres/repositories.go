package postgres

import "database/sql"

type Repositories struct {
	Users   *UserRepository
	Posts   *PostRepository
	Follows *FollowRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Users:   NewUserRepository(db),
		Posts:   NewPostRepository(db),
		Follows: NewFollowRepository(db),
	}
}
