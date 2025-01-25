package account

import (
	"github.com/luikyv/go-open-finance/internal/page"
)

type Storage struct {
	accountsMap map[string]Account
}

func NewStorage() *Storage {
	return &Storage{
		accountsMap: map[string]Account{},
	}
}

func (s *Storage) save(acc Account) {
	s.accountsMap[acc.ID] = acc
}

func (s *Storage) account(id string) Account {
	return s.accountsMap[id]
}

func (s *Storage) transactions(accID string, pag page.Pagination, filter transactionFilter) page.Page[Transaction] {
	acc := s.account(accID)
	var trs []Transaction
	for _, tr := range acc.Transactions {
		if tr.DateTime.Before(filter.from.Time) || tr.DateTime.After(filter.to.Time) {
			continue
		}

		trs = append(trs, tr)
	}

	return page.Paginate(trs, pag)
}
