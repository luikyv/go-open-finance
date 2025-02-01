package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/oidc"
)

func AuthPermissions(next http.Handler, consentService consent.Service, permissions ...consent.Permission) http.Handler {
	return authPermissions(next, consentService, false, permissions...)
}

func AuthPermissionsWithPagination(next http.Handler, consentService consent.Service, permissions ...consent.Permission) http.Handler {
	return authPermissions(next, consentService, true, permissions...)
}

func authPermissions(next http.Handler, consentService consent.Service, pagination bool, permissions ...consent.Permission) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		consent, err := consentService.Consent(ctx, ctx.Value(api.CtxKeyConsentID).(string))
		if err != nil {
			slog.DebugContext(r.Context(), "the token is not active")
			err := api.NewError("UNAUTHORISED", http.StatusUnauthorized, "invalid token")
			if pagination {
				err = err.WithPagination()
			}
			api.WriteError(w, err)
			return
		}

		if !consent.IsAuthorized() {
			slog.DebugContext(r.Context(), "the consent is not authorized")
			err := api.NewError("INVALID_STATUS", http.StatusUnauthorized, "the consent is not authorized")
			if pagination {
				err = err.WithPagination()
			}
			api.WriteError(w, err)
			return
		}

		if !consent.HasPermissions(permissions) {
			slog.DebugContext(r.Context(), "the consent doesn't have the required permissions")
			err := api.NewError("INVALID_STATUS", http.StatusForbidden, "the consent is missing permissions")
			if pagination {
				err = err.WithPagination()
			}
			api.WriteError(w, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthScopes(next http.Handler, op provider.Provider, scopes ...goidc.Scope) http.Handler {
	return authScopes(next, op, false, scopes...)
}

func AuthScopesWithPagination(next http.Handler, op provider.Provider, scopes ...goidc.Scope) http.Handler {
	return authScopes(next, op, true, scopes...)
}

func authScopes(next http.Handler, op provider.Provider, pagination bool, scopes ...goidc.Scope) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenInfo, err := op.TokenInfoFromRequest(w, r)
		if err != nil {
			slog.DebugContext(r.Context(), "the token is not active")
			err := api.NewError("UNAUTHORISED", http.StatusUnauthorized, "invalid token")
			if pagination {
				err = err.WithPagination()
			}
			api.WriteError(w, err)
			return
		}

		tokenScopes := strings.Split(tokenInfo.Scopes, " ")
		if !areScopesValid(scopes, tokenScopes) {
			slog.DebugContext(r.Context(), "invalid scopes", slog.String("token_scopes", tokenInfo.Scopes))
			err := api.NewError("UNAUTHORISED", http.StatusUnauthorized, "token missing scopes")
			if pagination {
				err = err.WithPagination()
			}
			api.WriteError(w, err)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, api.CtxKeyClientID, tokenInfo.ClientID)
		ctx = context.WithValue(ctx, api.CtxKeySubject, tokenInfo.Subject)
		if consentID, ok := ConsentID(tokenInfo.Scopes); ok {
			ctx = context.WithValue(ctx, api.CtxKeyConsentID, consentID)
		}
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// areScopesValid verifies every scope in requiredScopes has a match among
// scopes.
// scopes can have more scopes than the defined at requiredScopes, but the
// contrary results in false.
func areScopesValid(requiredScopes []goidc.Scope, scopes []string) bool {
	for _, requiredScope := range requiredScopes {
		if !isScopeValid(requiredScope, scopes) {
			return false
		}
	}
	return true
}

// isScopeValid verifies if requireScope has a match in scopes.
func isScopeValid(requiredScope goidc.Scope, scopes []string) bool {
	for _, scope := range scopes {
		if requiredScope.Matches(scope) {
			return true
		}
	}

	return false
}

func ConsentID(scopes string) (string, bool) {
	for _, s := range strings.Split(scopes, " ") {
		if oidc.ScopeConsentID.Matches(s) {
			return strings.Replace(s, "consent:", "", 1), true
		}
	}
	return "", false
}
