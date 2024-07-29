package services

import (
	"context"

	"github.com/luikyv/go-opf/gopf/constants"
	"github.com/luikyv/go-opf/gopf/models"
	"github.com/luikyv/go-opf/gopf/repositories"
)

type Consent struct {
	userService User
	consentRepo repositories.Consent
}

func NewConsent(userService User, consentRepo repositories.Consent) Consent {
	return Consent{
		userService: userService,
		consentRepo: consentRepo,
	}
}

func (service Consent) Authorize(
	ctx context.Context,
	consentID string,
	permissions ...models.ConsentPermission,
) models.OPFError {

	consent, err := service.get(ctx, consentID)
	if err != nil {
		return err
	}

	if consent.Status != models.ConsentStatusAwaitingAuthorisation {
		return models.NewOPFError(constants.ErrorInvalidRequest, "consent in invalid status")
	}

	consent.Status = models.ConsentStatusAuthorised
	consent.Permissions = permissions
	return service.save(ctx, consent)
}

func (service Consent) CreateForClient(
	ctx context.Context,
	consent models.Consent,
	caller models.CallerInfo,
) models.OPFError {
	if err := validatePermissions(consent.Permissions); err != nil {
		return err
	}
	return service.save(ctx, consent)
}

func (service Consent) GetForClient(
	ctx context.Context,
	id string,
	caller models.CallerInfo,
) (
	models.Consent,
	models.OPFError,
) {
	consent, err := service.get(ctx, id)
	if err != nil {
		return models.Consent{}, err
	}

	if consent.ClientId != caller.ClientID {
		return models.Consent{}, models.NewOPFError(constants.ErrorUnauthorized, "client not authorized")
	}

	return consent, nil
}

func (service Consent) RejectForClient(
	ctx context.Context,
	id string,
	caller models.CallerInfo,
) models.OPFError {
	consent, err := service.GetForClient(ctx, id, caller)
	if err != nil {
		return err
	}

	if consent.Status == models.ConsentStatusRejected {
		return models.NewOPFError(constants.ErrorInvalidOperation, "consent is already rejected")
	}

	consent.Status = models.ConsentStatusRejected
	return service.save(ctx, consent)
}

func (service Consent) Reject(
	ctx context.Context,
	consentID string,
) models.OPFError {

	consent, err := service.get(ctx, consentID)
	if err != nil {
		return err
	}

	consent.Status = models.ConsentStatusRejected
	return service.save(ctx, consent)
}

func (service Consent) save(
	ctx context.Context,
	consent models.Consent,
) models.OPFError {
	if err := service.consentRepo.Save(ctx, consent); err != nil {
		return models.NewOPFError(constants.ErrorInternalError, err.Error())
	}
	return nil
}

func (service Consent) get(ctx context.Context, id string) (models.Consent, models.OPFError) {
	consent, err := service.consentRepo.Get(ctx, id)
	if err != nil {
		return models.Consent{}, models.NewOPFError(constants.ErrorNotFound, err.Error())
	}

	if consent.Status != models.ConsentStatusRejected && consent.IsExpired() {
		consent.Status = models.ConsentStatusRejected
		consent.Status = models.ConsentStatusRejected
		if err := service.save(ctx, consent); err != nil {
			return models.Consent{}, models.NewOPFError(constants.ErrorInternalError, err.Error())
		}
	}

	return consent, nil
}
