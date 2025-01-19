package consent

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/timex"
)

var (
	errAccessNotAllowed                       = errors.New("access to consent is not allowed")
	errInvalidPermissionGroup                 = errors.New("the requested permission groups are invalid")
	errInvalidExpiration                      = errors.New("the expiration date time is invalid")
	errPersonalAndBusinessPermissionsTogether = errors.New("cannot request personal and business permissions together")
	errAlreadyRejected                        = errors.New("the consent is already rejected")
)

type Service struct {
	storage *Storage
}

func NewService(st *Storage) Service {
	return Service{
		storage: st,
	}
}

func (s Service) Authorize(ctx context.Context, id string, permissions ...Permission) error {

	slog.DebugContext(ctx, "trying to authorize consent", slog.String("consent_id", id))

	consent, err := s.Fetch(ctx, id)
	if err != nil {
		return err
	}

	if !consent.IsAwaitingAuthorization() {
		slog.DebugContext(ctx, "cannot authorize a consent that is not awaiting authorization", slog.Any("status", consent.Status))
		return errors.New("invalid consent status")
	}

	slog.InfoContext(ctx, "authorizing consent", slog.String("consent_id", id))
	consent.Status = StatusAuthorized
	consent.Permissions = permissions
	return s.save(ctx, consent)
}

func (s Service) Fetch(ctx context.Context, id string) (Consent, error) {
	c, err := s.storage.fetch(ctx, id)
	if err != nil {
		return Consent{}, err
	}

	// TODO: Should I remove this from here?
	if ctx.Value(api.CtxKeyClientID) != nil && ctx.Value(api.CtxKeyClientID) != c.ClientID {
		return Consent{}, errAccessNotAllowed
	}

	if err := s.modify(ctx, &c); err != nil {
		return Consent{}, err
	}

	return c, nil
}

func (s Service) Reject(ctx context.Context, id string, info RejectionInfo) error {
	c, err := s.Fetch(ctx, id)
	if err != nil {
		return err
	}
	if c.Status == StatusRejected {
		return errAlreadyRejected
	}

	c.Status = StatusRejected
	c.RejectionInfo = &info
	return s.save(ctx, c)
}

func (s Service) delete(ctx context.Context, id string) error {
	c, err := s.Fetch(ctx, id)
	if err != nil {
		return err
	}

	c.RejectionInfo = &RejectionInfo{
		RejectedBy: RejectedByUser,
		Reason:     RejectionReasonCustomerManuallyRejected,
	}
	if c.IsAuthorized() {
		c.RejectionInfo.Reason = RejectionReasonCustomerManuallyRevoked
	}
	c.Status = StatusRejected

	return s.save(ctx, c)
}

func (s Service) create(ctx context.Context, c Consent) error {
	if err := validate(c); err != nil {
		return err
	}

	return s.save(ctx, c)
}

// modify will evaluated the consent information and modify it to be compliant.
func (s Service) modify(ctx context.Context, consent *Consent) error {
	consentWasModified := false

	// Reject the consent if the time awaiting the user authorization has elapsed.
	if consent.HasAuthExpired() {
		slog.DebugContext(ctx, "consent awaiting authorization for too long, moving to rejected")
		consent.Status = StatusRejected
		consent.RejectionInfo = &RejectionInfo{
			RejectedBy: RejectedByUser,
			Reason:     RejectionReasonConsentExpired,
		}
		consentWasModified = true
	}

	// Reject the consent if it reached the expiration.
	if consent.IsExpired() {
		slog.DebugContext(ctx, "consent reached expiration, moving to rejected")
		consent.Status = StatusRejected
		consent.RejectionInfo = &RejectionInfo{
			RejectedBy: RejectedByASPSP,
			Reason:     RejectionReasonConsentMaxDateReached,
		}
		consentWasModified = true
	}

	if consentWasModified {
		slog.DebugContext(ctx, "the consent was modified")
		if err := s.save(ctx, *consent); err != nil {
			return err
		}
	}

	return nil
}

func (s Service) save(ctx context.Context, c Consent) error {
	return s.storage.save(ctx, c)
}

func validate(c Consent) error {
	if err := validatePermissions(c.Permissions); err != nil {
		return err
	}

	now := timex.Now()
	if c.ExpirationDateTime != nil && c.ExpirationDateTime.After(now.AddDate(1, 0, 0)) {
		return errInvalidExpiration
	}

	if c.ExpirationDateTime != nil && c.ExpirationDateTime.Before(now) {
		return errInvalidExpiration
	}

	return nil
}

func validatePermissions(requestedPermissions []Permission) error {

permissionsLoop:
	// Make sure if a permission is requested, at least one group of permissions
	// containing it is requested as well.
	for _, requestedPermission := range requestedPermissions {
		for _, group := range PermissionGroups {

			if slices.Contains(group, requestedPermission) && containsAll(requestedPermissions, group...) {
				continue permissionsLoop
			}

		}

		// Return an error if there is no group that contains requestedPermission
		// and is fully present in requestedPermissions.
		return errInvalidPermissionGroup
	}

	return validatePersonalAndBusinessPermissions(requestedPermissions)
}

func validatePersonalAndBusinessPermissions(requestedPermissions []Permission) error {
	isPersonal := containsAny(requestedPermissions,
		PermissionCustomersPersonalIdentificationsRead,
		PermissionCustomersPersonalAdittionalInfoRead,
	)
	isBusiness := containsAny(requestedPermissions,
		PermissionCustomersBusinessIdentificationsRead,
		PermissionCustomersBusinessAdittionalInfoRead,
	)

	if isPersonal && isBusiness {
		return errPersonalAndBusinessPermissionsTogether
	}

	return nil
}

func containsAll[T comparable](superSet []T, subSet ...T) bool {
	for _, t := range subSet {
		if !slices.Contains(superSet, t) {
			return false
		}
	}

	return true
}

func containsAny[T comparable](slice1 []T, slice2 ...T) bool {
	for _, t := range slice2 {
		if slices.Contains(slice1, t) {
			return true
		}
	}

	return false
}
