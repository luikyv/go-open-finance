package resource

import (
	"context"

	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/mock"
	"github.com/luikyv/go-open-finance/internal/page"
)

type Service struct {
	consentService consent.Service
}

func NewService(consentService consent.Service) Service {
	return Service{
		consentService: consentService,
	}
}

func (s Service) resources(ctx context.Context, consentID string, pag page.Pagination) (page.Page[Resource], error) {
	c, err := s.consentService.Consent(ctx, consentID)
	if err != nil {
		return page.Page[Resource]{}, err
	}

	var rs []Resource
	if c.AccountID != "" {
		r := Resource{
			ID:     c.AccountID,
			Type:   TypeAccount,
			Status: StatusAvailable,
		}
		if c.UserCPF == mock.CPFWithJointAccount && mock.IsJointAccountPendingAuth(c.CreationDateTime) {
			r.Status = StatusPendingAuthorization
		}
		rs = append(rs, r)
	}

	return page.Paginate(rs, pag), nil
}
