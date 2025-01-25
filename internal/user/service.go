package user

import (
	"context"
	"errors"
)

var (
	errUserNotFound = errors.New("user not found")
)

type Service struct {
	storage *Storage
}

func NewService(storage *Storage) Service {
	return Service{
		storage: storage,
	}
}

func (s Service) Create(ctx context.Context, user User) {
	s.storage.create(ctx, user)
}

func (s Service) User(username string) (User, error) {
	user, err := s.storage.user(username)
	if err != nil {
		return User{}, errUserNotFound
	}

	return user, nil
}

func (s Service) UserByCPF(cpf string) (User, error) {
	user, err := s.storage.userByCPF(cpf)
	if err != nil {
		return User{}, errUserNotFound
	}

	return user, nil
}
