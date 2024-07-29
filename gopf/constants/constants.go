package constants

import (
	"net/http"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

const (
	Namespace                = "urn:gopf"
	DatabaseSchema           = "gopf"
	DatabaseStringConnection = "mongodb://admin:password@localhost:27018"
	Port                     = ":80"
	Host                     = "https://gopf.localhost"
	MTLSHost                 = "https://matls-gopf.localhost"

	// API base paths.
	APIPrefixOIDC        = "/auth"
	BaseURLOIDC          = Host + APIPrefixOIDC
	APIPrefixOpenBanking = "/open-banking"
	APIPrefixConsentsV3  = "/consents/v3"
	BaseURLOpenBanking   = MTLSHost + APIPrefixOpenBanking
	BaseURLConsentsV3    = BaseURLOpenBanking + APIPrefixConsentsV3

	HeaderXFAPIInteractionID = "X-FAPI-Interaction-ID"

	// ACRs.
	ACROpenBankingLOA2 goidc.AuthenticationContextReference = "urn:brasil:openbanking:loa2"
	ACROpenBankingLOA3 goidc.AuthenticationContextReference = "urn:brasil:openbanking:loa3"

	RFC3339 = "2006-01-02T15:04:05Z"

	// Context keys.
	CtxKeySubject   = "ctx_subject"
	CtxKeyClientID  = "ctx_client_id"
	CtxKeyScopes    = "ctx_scopes"
	CtxKeyConsentID = "ctx_consent_id"
)

var (
	ScopeConsent = goidc.NewDynamicScope("consent", func(requestedScope string) bool {
		return strings.HasPrefix(requestedScope, "consent:")
	})
	ScopeConsents                    = goidc.NewScope("consents")
	ScopeCustomers                   = goidc.NewScope("customers")
	ScopeAccounts                    = goidc.NewScope("accounts")
	ScopeCreditCardAccounts          = goidc.NewScope("credit-cards-accounts")
	ScopeLoans                       = goidc.NewScope("loans")
	ScopeFinancings                  = goidc.NewScope("financings")
	ScopeUnarrangedAccountsOverdraft = goidc.NewScope("unarranged-accounts-overdraft")
	ScopeInvoiceFinancings           = goidc.NewScope("invoice-financings")
	ScopeBankFixedIncomes            = goidc.NewScope("bank-fixed-incomes")
	ScopeCreditFixedIncomes          = goidc.NewScope("credit-fixed-incomes")
	ScopeVariableIncomes             = goidc.NewScope("variable-incomes")
	ScopeTreasureTitles              = goidc.NewScope("treasure-titles")
	ScopeFunds                       = goidc.NewScope("funds")
	ScopeExchanges                   = goidc.NewScope("exchanges")
	ScopeResources                   = goidc.NewScope("resources")
)

type ErrorCode string

const (
	ErrorInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrorUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrorInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrorInvalidOperation ErrorCode = "INVALID_OPERATION"
	ErrorInvalidFAPIID    ErrorCode = "INVALID_INTERACTION_ID"
	ErrorNotFound         ErrorCode = "NOT_FOUND"
)

func (code ErrorCode) GetStatusCode() int {
	switch code {
	case ErrorUnauthorized:
		return http.StatusForbidden
	case ErrorInternalError:
		return http.StatusInternalServerError
	case ErrorNotFound:
		return http.StatusNotFound
	case ErrorInvalidOperation:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusBadRequest
	}
}
