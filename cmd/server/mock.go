package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/luikyv/go-open-finance/internal/timex"
	"github.com/luikyv/go-open-finance/internal/user"
)

const (
	mockBankCNPJ string = "18753655000104"
)

func loadMocks(
	userService user.Service,
) error {
	ctx := context.Background()

	var companyA = user.Company{
		CNPJ: "27737785000136",
	}
	var userBob = user.User{
		UserName:      "bob@mail.com",
		Email:         "bob@mail.com",
		CPF:           "78628584099",
		Name:          "Mr. Bob",
		AccountNumber: "12345678",
		CompanyCNPJs:  []string{companyA.CNPJ},
	}
	userService.Create(ctx, userBob)
	userService.AddPersonalIdentification(ctx, userBob.UserName, user.PersonalIdentification{
		ID:         uuid.NewString(),
		BrandName:  "MockBank",
		CivilName:  "Bob",
		SocialName: "Bob",
		BirthDate: timex.Date{
			Time: time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		MaritalStatus: user.MaritalStatusSOLTEIRO,
		Sex:           user.SexMale,
		CompanyCNPJ:   mockBankCNPJ,
		CPF:           userBob.CPF,
		Addresses: []user.Address{
			{
				IsMain:   true,
				Address:  "Av Paulista, 123",
				TownName: "São Paulo",
				PostCode: "00000000",
				Country:  "Brasil",
			},
		},
		Phones: []user.Phone{
			{
				IsMain:   true,
				Type:     user.PhoneTypeMobile,
				AreaCode: "11",
				Number:   "999999999",
			},
		},
		Emails: []user.Email{
			{
				IsMain: true,
				Email:  userBob.Email,
			},
		},
		UpdateDateTime: timex.DateTime{
			Time: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	})
	userService.SetPersonalQualification(ctx, userBob.UserName, user.PersonalQualifications{
		CompanyCNPJ:           mockBankCNPJ,
		Occupation:            user.OccupationOUTRO,
		OccupationDescription: "outra ocupação",
		UpdateDateTime: timex.DateTime{
			Time: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	})
	userService.SetPersonalFinancialRelations(ctx, userBob.UserName, user.PersonalFinancialRelations{
		ProductServiceTypes: []user.ProductServiceType{user.ProductServiceTypeCONTA_DEPOSITO_A_VISTA},
		Accounts: []user.Account{
			{
				CompeCode:  "000",
				Branch:     "0001",
				Number:     userBob.AccountNumber,
				CheckDigit: "1",
				Type:       user.AccountTypeCONTA_DEPOSITO_A_VISTA,
				SubType:    user.AccountSubTypeINDIVIDUAL,
			},
		},
		UpdateDateTime: timex.DateTime{
			Time: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		StartDateTime: timex.DateTime{
			Time: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	})

	return nil
}
