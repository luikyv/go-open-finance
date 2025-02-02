package consent

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"

	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/mock"
	"github.com/luikyv/go-open-finance/internal/page"
	"github.com/luikyv/go-open-finance/internal/timex"
)

var (
	errAccessNotAllowed                       = errors.New("access to consent is not allowed")
	errInvalidPermissionGroup                 = errors.New("the requested permission groups are invalid")
	errInvalidExpiration                      = errors.New("the expiration date time is invalid")
	errPersonalAndBusinessPermissionsTogether = errors.New("cannot request personal and business permissions together")
	errAlreadyRejected                        = errors.New("the consent is already rejected")
	errExtensionNotAllowed                    = errors.New("the consent is not allowed to be extended")
	errCannotExtendConsentNotAuthorized       = errors.New("the consent is not in the AUTHORISED status")
	errCannotExtendConsentForJointAccount     = errors.New("a consent created for a joint account cannot be extended")
)

func ID(scopes string) (string, bool) {
	for _, s := range strings.Split(scopes, " ") {
		if ScopeID.Matches(s) {
			return strings.Replace(s, "consent:", "", 1), true
		}
	}
	return "", false
}

type Service struct {
	storage Storage
}

func NewService(st Storage) Service {
	return Service{
		storage: st,
	}
}

func (s Service) Authorize(ctx context.Context, c Consent) error {

	slog.DebugContext(ctx, "trying to authorize consent", slog.String("consent_id", c.ID))

	if !c.IsAwaitingAuthorization() {
		slog.DebugContext(ctx, "cannot authorize a consent that is not awaiting authorization", slog.Any("status", c.Status))
		return errors.New("invalid consent status")
	}

	slog.InfoContext(ctx, "authorizing consent", slog.String("consent_id", c.ID))
	c.Status = StatusAuthorized
	c.StatusUpdateDateTime = timex.DateTimeNow()
	return s.save(ctx, c)
}

func (s Service) Consent(ctx context.Context, id string) (Consent, error) {
	c, err := s.storage.consent(ctx, id)
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
	c, err := s.Consent(ctx, id)
	if err != nil {
		return err
	}
	if c.Status == StatusRejected {
		return errAlreadyRejected
	}

	c.Status = StatusRejected
	c.StatusUpdateDateTime = timex.DateTimeNow()
	c.RejectionInfo = &info
	return s.save(ctx, c)
}

func (s Service) delete(ctx context.Context, id string) error {
	c, err := s.Consent(ctx, id)
	if err != nil {
		return err
	}

	info := RejectionInfo{
		RejectedBy: RejectedByUser,
		Reason:     RejectionReasonCustomerManuallyRejected,
	}
	if c.IsAuthorized() {
		info.Reason = RejectionReasonCustomerManuallyRevoked
	}

	return s.Reject(ctx, id, info)
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
		consent.StatusUpdateDateTime = timex.DateTimeNow()
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
		consent.StatusUpdateDateTime = timex.DateTimeNow()
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

func (s Service) extend(ctx context.Context, id string, ext Extension) (Consent, error) {
	c, err := s.Consent(ctx, id)
	if err != nil {
		return Consent{}, err
	}

	if c.UserCPF == mock.CPFWithJointAccount {
		return Consent{}, errCannotExtendConsentForJointAccount
	}

	if err := validateExtension(c, ext); err != nil {
		return Consent{}, err
	}

	ext.PreviousExpirationDateTime = c.ExpirationDateTime
	c.ExpirationDateTime = ext.ExpirationDateTime
	// The most recent extension must come first.
	c.Extensions = append([]Extension{ext}, c.Extensions...)
	if err := s.save(ctx, c); err != nil {
		return Consent{}, err
	}

	return c, nil
}

func (s Service) extensions(ctx context.Context, id string, pag page.Pagination) (page.Page[Extension], error) {
	c, err := s.Consent(ctx, id)
	if err != nil {
		return page.Page[Extension]{}, err
	}

	return page.Paginate(c.Extensions, pag), nil
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

func validateExtension(c Consent, ext Extension) error {
	if !c.IsAuthorized() {
		return errCannotExtendConsentNotAuthorized
	}

	if c.UserCPF != ext.UserCPF {
		return errExtensionNotAllowed
	}

	if c.BusinessCNPJ != "" && c.BusinessCNPJ != ext.BusinessCNPJ {
		return errExtensionNotAllowed
	}

	if ext.ExpirationDateTime == nil {
		return nil
	}

	now := timex.Now()
	if ext.ExpirationDateTime.Before(now) || ext.ExpirationDateTime.After(now.AddDate(1, 0, 0)) {
		return errInvalidExpiration
	}

	if c.ExpirationDateTime != nil && !ext.ExpirationDateTime.After(c.ExpirationDateTime.Time) {
		return errInvalidExpiration
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
