package creditcard

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
