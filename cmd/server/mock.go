package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/luikyv/go-open-finance/internal/account"
	"github.com/luikyv/go-open-finance/internal/customer"
	"github.com/luikyv/go-open-finance/internal/mock"
	"github.com/luikyv/go-open-finance/internal/timex"
	"github.com/luikyv/go-open-finance/internal/user"
)

var (
	random = rand.New(rand.NewSource(42))
)

func loadMocks(
	userService user.Service,
	customerService customer.Service,
	accountService account.Service,
) error {
	ctx := context.Background()

	if err := loadUserBob(ctx, userService, customerService, accountService); err != nil {
		return err
	}

	if err := loadUserAlice(ctx, userService, customerService, accountService); err != nil {
		return err
	}

	return nil
}

func loadUserBob(
	ctx context.Context,
	userService user.Service,
	customerService customer.Service,
	accountService account.Service,
) error {

	var u = user.User{
		UserName:  "bob@mail.com",
		Email:     "bob@mail.com",
		CPF:       "78628584099",
		Name:      "Mr. Bob",
		AccountID: uuid(),
	}
	userService.Create(ctx, u)

	customerService.AddPersonalIdentification(ctx, u.CPF, customer.PersonalIdentification{
		ID:            uuid(),
		BrandName:     "MockBank",
		CivilName:     "Bob",
		SocialName:    "Bob",
		BirthDate:     timex.NewDate(time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC)),
		MaritalStatus: customer.MaritalStatusSOLTEIRO,
		Sex:           customer.SexMale,
		CompanyCNPJ:   mock.MockBankCNPJ,
		CPF:           u.CPF,
		Addresses: []customer.Address{
			{
				IsMain:   true,
				Address:  "Av Paulista, 123",
				TownName: "São Paulo",
				PostCode: "00000000",
				Country:  "Brasil",
			},
		},
		Phones: []customer.Phone{
			{
				IsMain:   true,
				Type:     customer.PhoneTypeMobile,
				AreaCode: "11",
				Number:   "999999999",
			},
		},
		Emails: []customer.Email{
			{
				IsMain: true,
				Email:  u.Email,
			},
		},
		UpdateDateTime: timex.NewDateTime(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)),
	})
	customerService.SetPersonalQualification(ctx, u.CPF, customer.PersonalQualifications{
		CompanyCNPJ:           mock.MockBankCNPJ,
		Occupation:            customer.OccupationOUTRO,
		OccupationDescription: "outra ocupação",
		UpdateDateTime:        timex.NewDateTime(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)),
	})
	customerService.SetPersonalFinancialRelations(ctx, u.CPF, customer.PersonalFinancialRelations{
		ProductServiceTypes: []customer.ProductServiceType{customer.ProductServiceTypeCONTA_DEPOSITO_A_VISTA},
		Accounts: []customer.Account{
			{
				CompeCode:  "000",
				Branch:     "0001",
				Number:     "12345678",
				CheckDigit: "1",
				Type:       customer.AccountTypeCONTA_DEPOSITO_A_VISTA,
				SubType:    customer.AccountSubTypeINDIVIDUAL,
			},
		},
		UpdateDateTime: timex.NewDateTime(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)),
		StartDateTime:  timex.NewDateTime(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)),
	})

	acc := account.Account{
		ID:      u.AccountID,
		Number:  "53748219",
		Type:    account.TypeCheckingAccount,
		SubType: account.SubTypeIndividual,
		Balance: account.Balance{
			AvailableAmount:             "10000.00",
			BlockedAmount:               "0.00",
			AutomaticallyInvestedAmount: "0.00",
		},
	}

	for i := 0; i < 30; i++ {
		acc.Transactions = append(acc.Transactions, account.Transaction{
			ID:           uuid(),
			Status:       account.TransactionStatusCompleted,
			MovementType: account.MovementTypeCredit,
			Name:         fmt.Sprintf("Transaction %d", 30-i),
			Type:         account.TransactionTypePix,
			Amount:       "100.00",
			DateTime:     timex.NewDateTime(timex.DateTimeNow().Add(-time.Duration(i) * time.Hour)),
		})
	}

	for i := 0; i < 12; i++ {
		acc.Transactions = append(acc.Transactions, account.Transaction{
			ID:           uuid(),
			Status:       account.TransactionStatusCompleted,
			MovementType: account.MovementTypeCredit,
			Name:         fmt.Sprintf("Monthly Transaction %d", 12-i),
			Type:         account.TransactionTypePix,
			Amount:       "100.00",
			DateTime:     timex.NewDateTime(timex.DateTimeNow().AddDate(0, -i, 0)),
		})
	}

	accountService.Set(u.CPF, acc)

	return nil
}

func loadUserAlice(
	ctx context.Context,
	userService user.Service,
	_ customer.Service,
	accountService account.Service,
) error {
	var u = user.User{
		UserName:  "alice@mail.com",
		Email:     "alice@mail.com",
		CPF:       mock.CPFWithJointAccount,
		Name:      "Ms. Alice",
		AccountID: uuid(),
	}
	userService.Create(ctx, u)

	accountService.Set(u.CPF, account.Account{
		ID:      u.AccountID,
		Number:  "75690055",
		Type:    account.TypeCheckingAccount,
		SubType: account.SubTypeJointSimple,
		Balance: account.Balance{
			AvailableAmount:             "10000.00",
			BlockedAmount:               "0.00",
			AutomaticallyInvestedAmount: "0.00",
		},
		Transactions: []account.Transaction{
			{
				ID:           uuid(),
				Status:       account.TransactionStatusCompleted,
				MovementType: account.MovementTypeCredit,
				Name:         "First Transaction",
				Type:         account.TransactionTypePix,
				Amount:       "100.00",
				DateTime:     timex.NewDateTime(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
		},
	})

	return nil
}

// uuid generates a UUID-like string using a seeded random generator.
func uuid() string {
	b := make([]byte, 16)
	random.Read(b)
	return hex.EncodeToString(b[:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:])
}
