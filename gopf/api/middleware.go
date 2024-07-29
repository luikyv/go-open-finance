package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luikyv/go-opf/gopf/constants"
)

// FAPIHeaderMiddleware verifies if the FAPI interaction ID was sent and that it is valid.
// In case of a valid ID, it returns the same value in the response. Otherwise a random value is sent.
func FAPIHeaderMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		interactionID := ctx.GetHeader(constants.HeaderXFAPIInteractionID)
		if _, err := uuid.Parse(interactionID); err != nil {
			ctx.Abort()
			writeError(ctx, constants.ErrorInvalidFAPIID, "invalid interaction id")
		}

		ctx.Header(constants.HeaderXFAPIInteractionID, interactionID)
	}
}
