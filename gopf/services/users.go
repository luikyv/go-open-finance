package services

import (
	"github.com/luikyv/go-opf/gopf/models"
	"github.com/luikyv/go-opf/gopf/repositories"
)

type User struct {
	repo repositories.User
}

func NewUser(repo repositories.User) User {
	return User{
		repo: repo,
	}
}

func (service User) Get(username string) (models.User, models.OPFError) {
	return service.repo.Get(username)
}
