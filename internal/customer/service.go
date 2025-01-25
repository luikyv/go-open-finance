package customer

import (
	"context"

	"github.com/luikyv/go-open-finance/internal/page"
)

type Service struct {
	storage *Storage
}

func NewService(storage *Storage) Service {
	return Service{
		storage: storage,
	}
}

func (s Service) AddPersonalIdentification(_ context.Context, sub string, id PersonalIdentification) {
	s.storage.addPersonalIdentification(sub, id)
}

func (s Service) personalIdentifications(_ context.Context, sub string, p page.Pagination) page.Page[PersonalIdentification] {
	identifications := s.storage.personalIdentifications(sub)
	return page.Paginate(identifications, p)
}

func (s Service) SetPersonalQualification(_ context.Context, sub string, q PersonalQualifications) {
	s.storage.setPersonalQualification(sub, q)
}

func (s Service) personalQualifications(_ context.Context, sub string) PersonalQualifications {
	return s.storage.personalQualifications(sub)
}

func (s *Service) SetPersonalFinancialRelations(_ context.Context, sub string, fr PersonalFinancialRelations) {
	s.storage.setPersonalFinancialRelations(sub, fr)
}

func (s *Service) personalFinancialRelations(_ context.Context, sub string) PersonalFinancialRelations {
	return s.storage.personalFinancialRelations(sub)
}
