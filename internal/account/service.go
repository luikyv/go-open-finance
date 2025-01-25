package account

import (
	"context"
	"errors"
	"time"

	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/mock"
	"github.com/luikyv/go-open-finance/internal/page"
	"github.com/luikyv/go-open-finance/internal/timex"
)

var (
	errAccountNotAllowed                = errors.New("the account was not consented")
	errJointAccountPendingAuthorization = errors.New("the account was not authorized by all users")
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

func (s Service) Set(userID string, acc Account) {
	acc.UserID = userID
	s.storage.save(acc)
}

func (s Service) accounts(ctx context.Context, consentID string, pag page.Pagination) (page.Page[Account], error) {
	c, err := s.consentService.Consent(ctx, consentID)
	if err != nil {
		return page.Page[Account]{}, err
	}

	if c.UserCPF == mock.CPFWithJointAccount && c.CreationDateTime.Time.Before(timex.Now().Add(3*time.Minute)) {
		return page.Paginate([]Account{}, pag), nil
	}

	return page.Paginate([]Account{s.storage.account(c.AccountID)}, pag), nil
}

func (s Service) account(ctx context.Context, accID, consentID string) (Account, error) {
	c, err := s.consentService.Consent(ctx, consentID)
	if err != nil {
		return Account{}, err
	}

	if accID != c.AccountID {
		return Account{}, errAccountNotAllowed
	}

	if c.UserCPF == mock.CPFWithJointAccount && mock.IsJointAccountPendingAuth(c.CreationDateTime) {
		return Account{}, errJointAccountPendingAuthorization
	}

	return s.storage.account(accID), nil
}

func (s Service) transactions(
	ctx context.Context,
	accID, consentID string,
	pag page.Pagination,
	filter transactionFilter,
) (
	page.Page[Transaction],
	error,
) {
	consent, err := s.consentService.Consent(ctx, consentID)
	if err != nil {
		return page.Page[Transaction]{}, err
	}

	if accID != consent.AccountID {
		return page.Page[Transaction]{}, errAccountNotAllowed
	}
	return s.storage.transactions(accID, pag, filter), nil
}
