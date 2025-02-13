package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/luikyv/go-open-finance/internal/api"
)

const headerXFAPIInteractionID = "X-FAPI-Interaction-ID"

func FAPIID(next http.Handler) http.Handler {
	return fapiID(next, false)
}

func FAPIIDWithPagination(next http.Handler) http.Handler {
	return fapiID(next, true)
}

func fapiID(next http.Handler, pagination bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		interactionID := r.Header.Get(headerXFAPIInteractionID)
		// Verify if the interaction ID is valid, return a new value if not.
		if _, err := uuid.Parse(interactionID); err != nil {
			w.Header().Add(headerXFAPIInteractionID, uuid.NewString())
			err := api.NewError("INVALID_INTERACTION_ID", http.StatusBadRequest, "The fapi interaction id is missing or invalid")
			if pagination {
				err = err.WithPagination()
			}
			api.WriteError(w, err)
			return
		}

		// Return the same interaction ID in the response.
		w.Header().Add(headerXFAPIInteractionID, interactionID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, api.CtxKeyInteractionID, interactionID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
