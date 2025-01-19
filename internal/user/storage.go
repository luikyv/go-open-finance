package user

import (
	"context"
)

type Storage struct {
	users                         []User
	personalIdentificationsMap    map[string][]PersonalIdentification
	personalQualificationsMap     map[string]PersonalQualifications
	personalFinancialRelationsMap map[string]PersonalFinancialRelations
}

func NewStorage() *Storage {
	return &Storage{
		users:                         []User{},
		personalIdentificationsMap:    make(map[string][]PersonalIdentification),
		personalQualificationsMap:     make(map[string]PersonalQualifications),
		personalFinancialRelationsMap: make(map[string]PersonalFinancialRelations),
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

func (s *Storage) addPersonalIdentification(sub string, identification PersonalIdentification) {
	s.personalIdentificationsMap[sub] = append(s.personalIdentificationsMap[sub], identification)
}

func (s *Storage) personalIdentifications(sub string) []PersonalIdentification {
	return s.personalIdentificationsMap[sub]
}

func (s *Storage) setPersonalQualification(sub string, q PersonalQualifications) {
	s.personalQualificationsMap[sub] = q
}

func (s *Storage) personalQualifications(sub string) PersonalQualifications {
	return s.personalQualificationsMap[sub]
}

func (s *Storage) setPersonalFinancialRelations(sub string, rels PersonalFinancialRelations) {
	s.personalFinancialRelationsMap[sub] = rels
}

func (s *Storage) personalFinancialRelations(sub string) PersonalFinancialRelations {
	return s.personalFinancialRelationsMap[sub]
}

func findFirst[T any](elements []T, condition func(t T) bool) (T, bool) {
	for _, e := range elements {
		if condition(e) {
			return e, true
		}
	}

	return *new(T), false
}
