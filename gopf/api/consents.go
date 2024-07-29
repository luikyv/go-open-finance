package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luikyv/go-oidc/pkg/goidcp"
	"github.com/luikyv/go-opf/gopf/constants"
	"github.com/luikyv/go-opf/gopf/models"
	"github.com/luikyv/go-opf/gopf/services"
)

func RouteConsentsV3(router gin.IRouter, provider *goidcp.Provider, service services.Consent) {
	consentsRouter := router.Group(constants.APIPrefixConsentsV3)

	handler := postConsentsV3Handler(service)
	consentsRouter.POST("/consents", protectedHandler(handler, provider, constants.ScopeConsents))

	handler = getConsentsV3Handler(service)
	consentsRouter.GET("/consents/:consent_id", protectedHandler(handler, provider, constants.ScopeConsents))

	handler = deleteConsentsV3Handler(service)
	consentsRouter.DELETE("/consents/:consent_id", protectedHandler(handler, provider, constants.ScopeConsents))
}

func postConsentsV3Handler(service services.Consent) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.ConsentRequestV3
		if err := ctx.ShouldBindJSON(&req); err != nil {
			writeError(ctx, constants.ErrorInvalidRequest, err.Error())
			return
		}

		caller := models.NewCallerInfo(ctx)
		consent := models.NewConsentV3(req, caller.ClientID)
		if err := service.CreateForClient(ctx, consent, caller); err != nil {
			writeOPFError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, consent.NewConsentResponseV3())
	}
}

func getConsentsV3Handler(service services.Consent) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		caller := models.NewCallerInfo(ctx)
		consentID := ctx.Param("consent_id")
		consent, err := service.GetForClient(ctx, consentID, caller)
		if err != nil {
			writeOPFError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, consent.NewConsentResponseV3())
	}
}

func deleteConsentsV3Handler(service services.Consent) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		caller := models.NewCallerInfo(ctx)
		consentID := ctx.Param("consent_id")
		if err := service.RejectForClient(ctx, consentID, caller); err != nil {
			writeOPFError(ctx, err)
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
