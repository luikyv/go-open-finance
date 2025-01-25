package user

import (
	"context"
)

type Storage struct {
	users []User
}

func NewStorage() *Storage {
	return &Storage{
		users: []User{},
	}
}

func (st *Storage) create(_ context.Context, user User) {
	st.users = append(st.users, user)
}

func (st *Storage) user(username string) (User, error) {
	user, ok := findFirst(st.users, func(user User) bool {
		return user.UserName == username
	})
	if !ok {
		return User{}, errUserNotFound
	}
	return user, nil
}

func (st *Storage) userByCPF(cpf string) (User, error) {
	user, ok := findFirst(st.users, func(user User) bool {
		return user.CPF == cpf
	})
	if !ok {
		return User{}, errUserNotFound
	}

	return user, nil
}

func findFirst[T any](elements []T, condition func(t T) bool) (T, bool) {
	for _, e := range elements {
		if condition(e) {
			return e, true
		}
	}

	return *new(T), false
}
