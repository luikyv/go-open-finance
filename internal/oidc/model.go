package oidc

import (
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

const (
	// ACRs.
	ACROpenBankingLOA2 goidc.ACR = "urn:brasil:openbanking:loa2"
	ACROpenBankingLOA3 goidc.ACR = "urn:brasil:openbanking:loa3"
)

var (
	ScopeOpenID    = goidc.ScopeOpenID
	ScopeConsentID = goidc.NewDynamicScope("consent", func(requestedScope string) bool {
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

var Scopes = []goidc.Scope{
	ScopeOpenID,
	ScopeConsentID,
	ScopeConsents,
	ScopeCustomers,
	ScopeAccounts,
	ScopeCreditCardAccounts,
	ScopeLoans,
	ScopeFinancings,
	ScopeUnarrangedAccountsOverdraft,
	ScopeInvoiceFinancings,
	ScopeBankFixedIncomes,
	ScopeCreditFixedIncomes,
	ScopeVariableIncomes,
	ScopeTreasureTitles,
	ScopeFunds,
	ScopeExchanges,
	ScopeResources,
}
