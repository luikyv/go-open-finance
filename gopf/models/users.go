package models

type User struct {
	UserName  string
	CPF       string
	Name      string
	Companies []string
}

type Company struct {
	Name string
	CNPJ string
}
