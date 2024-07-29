package repositories

import "github.com/luikyv/go-opf/gopf/models"

var companyA = models.Company{
	Name: "A Business",
	CNPJ: "27737785000136",
}

var userBob = models.User{
	UserName:  "bob@mail.com",
	CPF:       "78628584099",
	Name:      "Mr. Bob",
	Companies: []string{companyA.CNPJ},
}
