package consent

import (
	"context"
	"errors"
)

type Storage struct {
	consents map[string]Consent
}

func NewStorage() *Storage {
	return &Storage{
		consents: make(map[string]Consent),
	}
}

func (st Storage) save(_ context.Context, consent Consent) error {
	st.consents[consent.ID] = consent
	return nil
}

func (st Storage) fetch(_ context.Context, id string) (Consent, error) {
	c, ok := st.consents[id]
	if !ok {
		return Consent{}, errors.New("a consent with the informed id was not found")
	}

	return c, nil
}
