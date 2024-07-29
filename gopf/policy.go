package main

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-opf/gopf/constants"
	"github.com/luikyv/go-opf/gopf/models"
	"github.com/luikyv/go-opf/gopf/services"
)

const (
	paramConsentID   = "consent_id"
	paramPermissions = "permissions"
	paramConsentCPF  = "consent_cpf"
	paramUserID      = "user_id"
	paramStepID      = "step_id"

	stepIDSetUp      = "setup"
	stepIDLogin      = "login"
	stepIDConsent    = "consent"
	stepIDFinishFlow = "finish_flow"

	usernameFormParam = "username"
	passwordFormParam = "password"
	consentFormParam  = "consent"

	correctPassword = "pass"
)

type AuthnPage struct {
	BaseURL     string
	CallbackID  string
	Permissions []models.ConsentPermission
	Error       string
}

func authenticationFunc(userService services.User, consentService services.Consent) goidc.AuthnFunc {
	return func(ctx goidc.Context, session *goidc.AuthnSession) goidc.AuthnStatus {
		if _, ok := session.Store[paramStepID]; !ok {
			session.StoreParameter(paramStepID, stepIDSetUp)
		}

		if session.Parameter(paramStepID) == stepIDSetUp {
			if status := setUp(ctx, session, consentService); status != goidc.StatusSuccess {
				return status
			}
			session.StoreParameter(paramStepID, stepIDLogin)
		}

		if session.Parameter(paramStepID) == stepIDLogin {
			if status := login(ctx, session, userService); status != goidc.StatusSuccess {
				return status
			}
			session.StoreParameter(paramStepID, stepIDConsent)
		}

		if session.Parameter(paramStepID) == stepIDConsent {
			if status := consent(ctx, session, consentService); status != goidc.StatusSuccess {
				return status
			}
			session.StoreParameter(paramStepID, stepIDFinishFlow)
		}

		if session.Parameter(paramStepID) == stepIDFinishFlow {
			return finishFlow(ctx, session)
		}

		return goidc.StatusFailure
	}
}

func setUp(
	ctx goidc.Context,
	session *goidc.AuthnSession,
	consentService services.Consent,
) goidc.AuthnStatus {
	consentID := consentID(session.Scopes)
	if consentID == "" {
		session.SetRedirectError(goidc.ErrorCodeAccessDenied, "missing consent ID")
		return goidc.StatusFailure
	}

	caller := models.CallerInfo{
		ClientID: session.ClientID,
	}
	consent, err := consentService.GetForClient(ctx, consentID, caller)
	if err != nil {
		session.SetRedirectError(goidc.ErrorCodeAccessDenied, err.Error())
		return goidc.StatusFailure
	}

	session.StoreParameter(paramConsentID, consent.ID)
	permissions := ""
	for _, p := range consent.Permissions {
		permissions += fmt.Sprintf("%s ", p)
	}
	permissions = permissions[:len(permissions)-1]
	session.StoreParameter(paramPermissions, permissions)
	session.StoreParameter(paramConsentCPF, consent.UserCPF)
	return goidc.StatusSuccess
}

func login(
	ctx goidc.Context,
	session *goidc.AuthnSession,
	userService services.User,
) goidc.AuthnStatus {

	ctx.Request().ParseForm()

	username := ctx.Request().PostFormValue(usernameFormParam)
	if username == "" {
		ctx.Response().WriteHeader(http.StatusOK)
		tmpl, _ := template.ParseFiles("../templates/login.html")
		tmpl.Execute(ctx.Response(), AuthnPage{
			BaseURL:    constants.BaseURLOIDC,
			CallbackID: session.CallbackID,
		})
		return goidc.StatusInProgress
	}

	user, err := userService.Get(username)
	if err != nil {
		ctx.Response().WriteHeader(http.StatusOK)
		tmpl, _ := template.ParseFiles("../templates/login.html")
		tmpl.Execute(ctx.Response(), AuthnPage{
			BaseURL:    constants.BaseURLOIDC,
			CallbackID: session.CallbackID,
			Error:      "invalid username",
		})
		return goidc.StatusInProgress
	}

	password := ctx.Request().PostFormValue(passwordFormParam)
	if user.CPF != session.Parameter(paramConsentCPF) || password != correctPassword {
		ctx.Response().WriteHeader(http.StatusOK)
		tmpl, _ := template.ParseFiles("../templates/login.html")
		tmpl.Execute(ctx.Response(), AuthnPage{
			BaseURL:    constants.BaseURLOIDC,
			CallbackID: session.CallbackID,
			Error:      "invalid credentials",
		})
		return goidc.StatusInProgress
	}

	session.StoreParameter(paramUserID, username)
	return goidc.StatusSuccess
}

func consent(
	ctx goidc.Context,
	session *goidc.AuthnSession,
	consentService services.Consent,
) goidc.AuthnStatus {

	ctx.Request().ParseForm()

	var permissions []models.ConsentPermission
	for _, p := range strings.Split(session.Parameter(paramPermissions).(string), " ") {
		permissions = append(permissions, models.ConsentPermission(p))
	}
	isConsented := ctx.Request().PostFormValue(consentFormParam)
	if isConsented == "" {
		ctx.Response().WriteHeader(http.StatusOK)
		tmpl, _ := template.ParseFiles("../templates/consent.html")
		tmpl.Execute(ctx.Response(), AuthnPage{
			BaseURL:     constants.BaseURLOIDC,
			CallbackID:  session.CallbackID,
			Permissions: permissions,
		})
		return goidc.StatusInProgress
	}

	consentID := session.Parameter(paramConsentID).(string)

	if isConsented != "true" {
		consentService.Reject(ctx, consentID)
		session.SetRedirectError(goidc.ErrorCodeAccessDenied, "consent not granted")
		return goidc.StatusFailure
	}

	if err := consentService.Authorize(ctx, consentID, permissions...); err != nil {
		session.SetRedirectError(goidc.ErrorCodeAccessDenied, err.Error())
		return goidc.StatusFailure
	}
	return goidc.StatusSuccess
}

func finishFlow(
	ctx goidc.Context,
	session *goidc.AuthnSession,
) goidc.AuthnStatus {
	session.SetUserID(session.Parameter(paramUserID).(string))
	session.GrantScopes(session.Scopes)
	session.SetACRClaimIDToken(constants.ACROpenBankingLOA2)
	session.SetAuthTimeClaimIDToken(int(time.Now().Unix()))
	handleClaimsObject(ctx, session)

	return goidc.StatusSuccess
}

func handleClaimsObject(_ goidc.Context, session *goidc.AuthnSession) {
	if session.Claims == nil {
		return
	}

	if slices.Contains(session.Claims.IDTokenEssentials(), goidc.ClaimAuthenticationContextReference) {
		session.SetACRClaimIDToken(constants.ACROpenBankingLOA2)
	}

	if slices.Contains(session.Claims.UserInfoEssentials(), goidc.ClaimAuthenticationContextReference) {
		session.SetACRClaimUserInfo(constants.ACROpenBankingLOA2)
	}
}

func consentID(scopes string) string {
	for _, s := range strings.Split(scopes, " ") {
		if constants.ScopeConsent.Matches(s) {
			return strings.Replace(s, "consent:", "", 1)
		}
	}
	return ""
}
