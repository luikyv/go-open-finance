package repositories

import (
	"github.com/luikyv/go-opf/gopf/constants"
	"github.com/luikyv/go-opf/gopf/models"
)

type User struct {
	usersMap map[string]models.User
}

func NewUser() User {
	return User{
		usersMap: map[string]models.User{
			userBob.UserName: userBob,
		},
	}
}

func (repo User) Get(username string) (models.User, models.OPFError) {
	user, ok := repo.usersMap[username]
	if !ok {
		return models.User{}, models.NewOPFError(constants.ErrorNotFound, "user not found")
	}
	return user, nil
}
