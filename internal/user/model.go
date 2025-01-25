package user

import (
	"slices"
)

type User struct {
	UserName     string
	Email        string
	CPF          string
	Name         string
	AccountID    string
	CompanyCNPJs []string
}

func (u User) OwnsCompany(cnpj string) bool {
	return slices.Contains(u.CompanyCNPJs, cnpj)
}

type Company struct {
	Name string
	CNPJ string
}
