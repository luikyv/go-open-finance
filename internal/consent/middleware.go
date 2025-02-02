package consent

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/luikyv/go-open-finance/internal/api"
)

func PermissionMiddleware(next http.Handler, consentService Service, permissions ...Permission) http.Handler {
	return permissionMiddleware(next, consentService, false, permissions...)
}

func PermissionMiddlewareWithPagination(next http.Handler, consentService Service, permissions ...Permission) http.Handler {
	return permissionMiddleware(next, consentService, true, permissions...)
}

func permissionMiddleware(next http.Handler, consentService Service, pagination bool, permissions ...Permission) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		scopes := ctx.Value(api.CtxKeyScopes).(string)

		id, _ := ID(scopes)
		consent, err := consentService.Consent(ctx, id)
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

		ctx = context.WithValue(ctx, api.CtxKeyConsentID, id)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
