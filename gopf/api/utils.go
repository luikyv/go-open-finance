package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/goidcp"
	"github.com/luikyv/go-opf/gopf/constants"
	"github.com/luikyv/go-opf/gopf/models"
)

func writeError(ctx *gin.Context, errorCode constants.ErrorCode, description string) {
	ctx.JSON(
		errorCode.GetStatusCode(),
		models.NewResponseError(errorCode, description),
	)
}

func writeOPFError(ctx *gin.Context, err models.OPFError) {
	ctx.JSON(
		err.Code().GetStatusCode(),
		err.Response(),
	)
}

// protectedHandler returns a new handler that executes the informed handler if the request contains the right scopes.
func protectedHandler(
	handler gin.HandlerFunc,
	provider *goidcp.Provider,
	scopeSlice ...goidc.Scope,
) gin.HandlerFunc {
	scopes := goidc.Scopes(scopeSlice)
	return func(ctx *gin.Context) {
		tokenInfo := provider.TokenInfo(ctx.Request, ctx.Writer)
		if !tokenInfo.IsActive {
			writeError(ctx, constants.ErrorUnauthorized, "invalid token")
			return
		}

		tokenScopes := strings.Split(tokenInfo.Scopes, " ")
		if !scopes.AreContainedIn(tokenScopes) {
			writeError(ctx, constants.ErrorUnauthorized, "token missing scopes")
			return
		}

		// Add more context to the request.
		ctx.Set(constants.CtxKeySubject, tokenInfo.Subject)
		ctx.Set(constants.CtxKeyClientID, tokenInfo.ClientID)
		ctx.Set(constants.CtxKeyScopes, tokenInfo.Scopes)
		for _, tokenScope := range tokenScopes {
			if constants.ScopeConsent.Matches(tokenScope) {
				ctx.AddParam(constants.CtxKeyConsentID, strings.Replace(tokenScope, "consent:", "", 1))
			}
		}

		// Execute the protected handler.
		handler(ctx)
	}
}
