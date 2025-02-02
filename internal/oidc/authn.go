package oidc

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/timex"
	"github.com/luikyv/go-open-finance/internal/user"
)

func Policy(
	templatesDir, baseURL string,
	userService user.Service,
	consentService consent.Service,
) goidc.AuthnPolicy {

	loginTemplate := filepath.Join(templatesDir, "/login.html")
	consentTemplate := filepath.Join(templatesDir, "/consent.html")
	tmpl, err := template.ParseFiles(loginTemplate, consentTemplate)
	if err != nil {
		log.Fatal(err)
	}

	authenticator := authenticator{
		tmpl:           tmpl,
		baseURL:        baseURL,
		userService:    userService,
		consentService: consentService,
	}
	return goidc.NewPolicy(
		"main",
		func(r *http.Request, c *goidc.Client, as *goidc.AuthnSession) bool {
			as.StoreParameter(paramStepID, stepIDSetUp)
			return true
		},
		authenticator.authenticate,
	)
}

const (
	paramConsentID   = "consent_id"
	paramPermissions = "permissions"
	paramConsentCPF  = "consent_cpf"
	paramConsentCNPJ = "consent_cnpj"
	paramUserID      = "user_id"
	paramStepID      = "step_id"

	stepIDSetUp      = "setup"
	stepIDLogin      = "login"
	stepIDConsent    = "consent"
	stepIDFinishFlow = "finish_flow"

	usernameFormParam = "username"
	passwordFormParam = "password"
	loginFormParam    = "login"
	consentFormParam  = "consent"

	correctPassword = "pass"
)

type authnPage struct {
	CallbackID   string
	UserCPF      string
	BusinessCNPJ string
	Permissions  []consent.Permission
	Error        string
}

type authenticator struct {
	tmpl           *template.Template
	baseURL        string
	userService    user.Service
	consentService consent.Service
}

func (a authenticator) authenticate(w http.ResponseWriter, r *http.Request, session *goidc.AuthnSession) (goidc.AuthnStatus, error) {
	if session.StoredParameter(paramStepID) == stepIDSetUp {
		if status, err := a.setUp(r, session); status != goidc.StatusSuccess {
			return status, err
		}
		session.StoreParameter(paramStepID, stepIDLogin)
	}

	if session.StoredParameter(paramStepID) == stepIDLogin {
		if status, err := a.login(w, r, session); status != goidc.StatusSuccess {
			return status, err
		}
		session.StoreParameter(paramStepID, stepIDConsent)
	}

	if session.StoredParameter(paramStepID) == stepIDConsent {
		if status, err := a.grantConsent(w, r, session); status != goidc.StatusSuccess {
			return status, err
		}
		session.StoreParameter(paramStepID, stepIDFinishFlow)
	}

	if session.StoredParameter(paramStepID) == stepIDFinishFlow {
		return a.finishFlow(session)
	}

	return goidc.StatusFailure, errors.New("access denied")
}

func (a authenticator) setUp(r *http.Request, session *goidc.AuthnSession) (goidc.AuthnStatus, error) {
	consentID, ok := consent.ID(session.Scopes)
	if !ok {
		return goidc.StatusFailure, errors.New("missing consent ID")
	}

	consent, err := a.consentService.Consent(r.Context(), consentID)
	if err != nil {
		return goidc.StatusFailure, err
	}

	if !consent.IsAwaitingAuthorization() {
		return goidc.StatusFailure, errors.New("consent is not awaiting authorization")
	}

	user, err := a.userService.UserByCPF(consent.UserCPF)
	if err != nil {
		return goidc.StatusFailure, errors.New("the consent was created for an user that does not exist")
	}

	if consent.BusinessCNPJ != "" && !user.OwnsCompany(consent.BusinessCNPJ) {
		return goidc.StatusFailure, errors.New("the consent was created for a business that is not available to the logged user")
	}

	// Convert permissions to []string for joining.
	permissionsStr := make([]string, len(consent.Permissions))
	for i, permission := range consent.Permissions {
		permissionsStr[i] = string(permission)
	}

	// TODO: Remove this.
	consent.AccountID = user.AccountID

	session.StoreParameter(paramConsentID, consent.ID)
	session.StoreParameter(paramPermissions, strings.Join(permissionsStr, " "))
	session.StoreParameter(paramConsentCPF, consent.UserCPF)
	if consent.BusinessCNPJ != "" {
		session.StoreParameter(paramConsentCNPJ, consent.BusinessCNPJ)
	}
	return goidc.StatusSuccess, nil
}

func (a authenticator) login(w http.ResponseWriter, r *http.Request, session *goidc.AuthnSession) (goidc.AuthnStatus, error) {

	_ = r.ParseForm()

	isLogin := r.PostFormValue(loginFormParam)
	if isLogin == "" {
		return a.executeTemplate(w, "login.html", authnPage{
			CallbackID: session.CallbackID,
		})
	}

	if isLogin != "true" {
		consentID := session.StoredParameter(paramConsentID).(string)
		_ = a.consentService.Reject(r.Context(), consentID, consent.RejectionInfo{
			RejectedBy: consent.RejectedByUser,
			Reason:     consent.RejectionReasonCustomerManuallyRejected,
		})
		return goidc.StatusFailure, errors.New("consent not granted")
	}

	username := r.PostFormValue(usernameFormParam)
	user, err := a.userService.User(username)
	if err != nil {
		return a.executeTemplate(w, "login.html", authnPage{
			CallbackID: session.CallbackID,
			Error:      "invalid username",
		})
	}

	password := r.PostFormValue(passwordFormParam)
	if user.CPF != session.StoredParameter(paramConsentCPF) || password != correctPassword {
		return a.executeTemplate(w, "login.html", authnPage{
			CallbackID: session.CallbackID,
			Error:      "invalid credentials",
		})
	}

	session.StoreParameter(paramUserID, username)
	return goidc.StatusSuccess, nil
}

func (a authenticator) grantConsent(w http.ResponseWriter, r *http.Request, session *goidc.AuthnSession) (goidc.AuthnStatus, error) {

	_ = r.ParseForm()

	var permissions []consent.Permission
	for _, p := range strings.Split(session.StoredParameter(paramPermissions).(string), " ") {
		permissions = append(permissions, consent.Permission(p))
	}

	isConsented := r.PostFormValue(consentFormParam)
	if isConsented == "" {
		page := authnPage{
			CallbackID:  session.CallbackID,
			UserCPF:     session.StoredParameter(paramConsentCPF).(string),
			Permissions: permissions,
		}
		if cnpj := session.StoredParameter(paramConsentCNPJ); cnpj != nil {
			page.BusinessCNPJ = cnpj.(string)
		}
		return a.executeTemplate(w, "consent.html", page)
	}

	consentID := session.StoredParameter(paramConsentID).(string)

	if isConsented != "true" {
		_ = a.consentService.Reject(r.Context(), consentID, consent.RejectionInfo{
			RejectedBy: consent.RejectedByUser,
			Reason:     consent.RejectionReasonCustomerManuallyRejected,
		})
		return goidc.StatusFailure, errors.New("consent not granted")
	}

	c, err := a.consentService.Consent(r.Context(), consentID)
	if err != nil {
		return goidc.StatusFailure, err
	}
	u, err := a.userService.UserByCPF(c.UserCPF)
	if err != nil {
		return goidc.StatusFailure, err
	}

	if slices.ContainsFunc(c.Permissions, func(p consent.Permission) bool {
		return strings.HasPrefix(string(p), "ACCOUNTS_")
	}) {
		c.AccountID = u.AccountID
	}

	if err := a.consentService.Authorize(r.Context(), c); err != nil {
		return goidc.StatusFailure, err
	}
	return goidc.StatusSuccess, nil
}

func (a authenticator) finishFlow(session *goidc.AuthnSession) (goidc.AuthnStatus, error) {
	session.SetUserID(session.StoredParameter(paramUserID).(string))
	session.GrantScopes(session.Scopes)
	session.SetIDTokenClaimACR(ACROpenBankingLOA2)
	session.SetIDTokenClaimAuthTime(int(timex.Now().Unix()))

	if session.Claims != nil {
		if slices.Contains(session.Claims.IDTokenEssentials(), goidc.ClaimACR) {
			session.SetIDTokenClaimACR(ACROpenBankingLOA2)
		}

		if slices.Contains(session.Claims.UserInfoEssentials(), goidc.ClaimACR) {
			session.SetUserInfoClaimACR(ACROpenBankingLOA2)
		}
	}

	return goidc.StatusSuccess, nil
}

func (a authenticator) executeTemplate(
	w http.ResponseWriter,
	templateName string,
	params authnPage,
) (
	goidc.AuthnStatus,
	error,
) {
	type page struct {
		BaseURL string
		authnPage
	}
	w.WriteHeader(http.StatusOK)
	_ = a.tmpl.ExecuteTemplate(w, templateName, page{
		BaseURL:   a.baseURL,
		authnPage: params,
	})
	return goidc.StatusInProgress, nil
}
