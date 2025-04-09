package creditcard

import (
	"context"
	"errors"

	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/page"
)

var (
	errAccountNotAllowed = errors.New("the account was not consented")
)

type Service struct {
	storage        *Storage
	consentService consent.Service
}

func NewService(storage *Storage, consentService consent.Service) Service {
	return Service{
		storage:        storage,
		consentService: consentService,
	}
}

func (s Service) Add(userID string, acc Account) {
	acc.UserID = userID
	s.storage.save(acc)
}

func (s Service) accounts(ctx context.Context, consentID string, pag page.Pagination) (page.Page[Account], error) {
	c, err := s.consentService.Consent(ctx, consentID)
	if err != nil {
		return page.Page[Account]{}, err
	}

	return page.Paginate([]Account{s.storage.account(c.CreditAccountID)}, pag), nil
}

func (s Service) account(ctx context.Context, id, consentID string) (Account, error) {
	c, err := s.consentService.Consent(ctx, consentID)
	if err != nil {
		return Account{}, err
	}

	if id != c.CreditAccountID {
		return Account{}, errAccountNotAllowed
	}

	return s.storage.account(id), nil
}
