package services

import (
	"slices"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-opf/gopf/constants"
	"github.com/luikyv/go-opf/gopf/models"
)

func validatePermissions(requestedPermissions []models.ConsentPermission) models.OPFError {

	if !goidc.ContainsAll(models.ConsentPermissions, requestedPermissions...) {
		return models.NewOPFError(constants.ErrorInvalidRequest, "invalid permission")
	}

permissionsLoop:
	// Make sure if a permission is requested, at least one group of permissions containing it
	// is requested as well.
	for _, requestedPermission := range requestedPermissions {
		for _, group := range models.PermissionGroups {
			if slices.Contains(group, requestedPermission) && goidc.ContainsAll(requestedPermissions, group...) {
				continue permissionsLoop
			}
		}
		// Return an error if there is no group that contains requestedPermission and is fully present in requestedPermissions.
		return models.NewOPFError(constants.ErrorInvalidOperation, "cannot request a permission without all the others from its group")
	}
	return nil
}
