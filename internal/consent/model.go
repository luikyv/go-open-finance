package consent

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-open-finance/internal/timex"
)

const (
	maxTimeAwaitingAuthorizationSecs = 3600
	headerCustomerIPAddress          = "X-FAPI-Customer-IP-Address"
	headerCustomerUserAgent          = "X-Customer-User-Agent"
	defaultUserDocumentRelation      = "CPF"
	defaultBusinessDocumentRelation  = "CNPJ"
)

var (
	ScopeID = goidc.NewDynamicScope("consent", func(requestedScope string) bool {
		return strings.HasPrefix(requestedScope, "consent:")
	})
	Scope = goidc.NewScope("consents")
)

type Consent struct {
	ID            string         `bson:"_id"`
	Status        Status         `bson:"status"`
	UserCPF       string         `bson:"user_cpf"`
	BusinessCNPJ  string         `bson:"business_cnpj,omitempty"`
	Permissions   []Permission   `bson:"permissions"`
	RejectionInfo *RejectionInfo `bson:"rejection,omitempty"`
	Extensions    []Extension    `bson:"extensions,omitempty"`

	ClientID             string          `bson:"client_id"`
	CreationDateTime     timex.DateTime  `bson:"created_at"`
	StatusUpdateDateTime timex.DateTime  `bson:"status_updated_at"`
	ExpirationDateTime   *timex.DateTime `bson:"expires_at,omitempty"`

	// Resources consented by the user.
	AccountID       string `json:"account_id,omitempty"`
	CreditAccountID string `json:"credit_account_id,omitempty"`
}

// HasAuthExpired returns true if the status is [StatusAwaitingAuthorisation] and
// the max time awaiting authorization has elapsed.
func (c Consent) HasAuthExpired() bool {
	now := timex.Now()
	return c.IsAwaitingAuthorization() &&
		now.After(c.CreationDateTime.Add(timex.Second*maxTimeAwaitingAuthorizationSecs))
}

// IsExpired returns true if the status is [StatusAuthorised] and the consent
// reached the expiration date.
func (c Consent) IsExpired() bool {
	if c.ExpirationDateTime == nil {
		return false
	}
	now := timex.Now()
	return c.Status == StatusAuthorized && now.After(c.ExpirationDateTime.Time)
}

func (c Consent) IsAuthorized() bool {
	return c.Status == StatusAuthorized
}

func (c Consent) IsAwaitingAuthorization() bool {
	return c.Status == StatusAwaitingAuthorization
}

func (c Consent) HasPermissions(permissions []Permission) bool {
	for _, p := range permissions {
		if !slices.Contains(c.Permissions, p) {
			return false
		}
	}

	return true
}

type Status string

const (
	StatusAwaitingAuthorization Status = "AWAITING_AUTHORISATION"
	StatusAuthorized            Status = "AUTHORISED"
	StatusRejected              Status = "REJECTED"
)

type Permission string

const (
	PermissionAccountsBalanceRead                                 Permission = "ACCOUNTS_BALANCES_READ"
	PermissionAccountsOverdraftLimitsRead                         Permission = "ACCOUNTS_OVERDRAFT_LIMITS_READ"
	PermissionAccountsRead                                        Permission = "ACCOUNTS_READ"
	PermissionAccountsTransactionsRead                            Permission = "ACCOUNTS_TRANSACTIONS_READ"
	PermissionBankFixedIncomesRead                                Permission = "BANK_FIXED_INCOMES_READ"
	PermissionCreditCardsAccountsBillsRead                        Permission = "CREDIT_CARDS_ACCOUNTS_BILLS_READ"
	PermissionCreditCardsAccountsBillsTransactionsRead            Permission = "CREDIT_CARDS_ACCOUNTS_BILLS_TRANSACTIONS_READ"
	PermissionCreditCardsAccountsLimitsRead                       Permission = "CREDIT_CARDS_ACCOUNTS_LIMITS_READ"
	PermissionCreditCardsAccountsRead                             Permission = "CREDIT_CARDS_ACCOUNTS_READ"
	PermissionCreditCardsAccountsTransactionsRead                 Permission = "CREDIT_CARDS_ACCOUNTS_TRANSACTIONS_READ"
	PermissionCreditFixedIncomesRead                              Permission = "CREDIT_FIXED_INCOMES_READ"
	PermissionCustomersBusinessAdittionalInfoRead                 Permission = "CUSTOMERS_BUSINESS_ADITTIONALINFO_READ"
	PermissionCustomersBusinessIdentificationsRead                Permission = "CUSTOMERS_BUSINESS_IDENTIFICATIONS_READ"
	PermissionCustomersPersonalAdittionalInfoRead                 Permission = "CUSTOMERS_PERSONAL_ADITTIONALINFO_READ"
	PermissionCustomersPersonalIdentificationsRead                Permission = "CUSTOMERS_PERSONAL_IDENTIFICATIONS_READ"
	PermissionExchangesRead                                       Permission = "EXCHANGES_READ"
	PermissionFinancingsPaymentsRead                              Permission = "FINANCINGS_PAYMENTS_READ"
	PermissionFinancingsRead                                      Permission = "FINANCINGS_READ"
	PermissionFinancingsScheduledInstalmentsRead                  Permission = "FINANCINGS_SCHEDULED_INSTALMENTS_READ"
	PermissionFinancingsWarrantiesRead                            Permission = "FINANCINGS_WARRANTIES_READ"
	PermissionFundsRead                                           Permission = "FUNDS_READ"
	PermissionInvoiceFinancingsPaymentsRead                       Permission = "INVOICE_FINANCINGS_PAYMENTS_READ"
	PermissionInvoiceFinancingsRead                               Permission = "INVOICE_FINANCINGS_READ"
	PermissionInvoiceFinancingsScheduledInstalmentsRead           Permission = "INVOICE_FINANCINGS_SCHEDULED_INSTALMENTS_READ"
	PermissionInvoiceFinancingsWarrantiesRead                     Permission = "INVOICE_FINANCINGS_WARRANTIES_READ"
	PermissionLoansPaymentsRead                                   Permission = "LOANS_PAYMENTS_READ"
	PermissionLoansRead                                           Permission = "LOANS_READ"
	PermissionLoansScheduledInstalmentsRead                       Permission = "LOANS_SCHEDULED_INSTALMENTS_READ"
	PermissionLoansWarrantiesRead                                 Permission = "LOANS_WARRANTIES_READ"
	PermissionResourcesRead                                       Permission = "RESOURCES_READ"
	PermissionTreasureTitlesRead                                  Permission = "TREASURE_TITLES_READ"
	PermissionUnarrangedAccountsOverdraftPaymentsRead             Permission = "UNARRANGED_ACCOUNTS_OVERDRAFT_PAYMENTS_READ"
	PermissionUnarrangedAccountsOverdraftRead                     Permission = "UNARRANGED_ACCOUNTS_OVERDRAFT_READ"
	PermissionUnarrangedAccountsOverdraftScheduledInstalmentsRead Permission = "UNARRANGED_ACCOUNTS_OVERDRAFT_SCHEDULED_INSTALMENTS_READ"
	PermissionUnarrangedAccountsOverdraftWarrantiesRead           Permission = "UNARRANGED_ACCOUNTS_OVERDRAFT_WARRANTIES_READ"
	PermissionVariableIncomesRead                                 Permission = "VARIABLE_INCOMES_READ"
)

var Permissions = []Permission{
	PermissionAccountsBalanceRead,
	PermissionAccountsOverdraftLimitsRead,
	PermissionAccountsRead,
	PermissionAccountsTransactionsRead,
	PermissionBankFixedIncomesRead,
	PermissionCreditCardsAccountsBillsRead,
	PermissionCreditCardsAccountsBillsTransactionsRead,
	PermissionCreditCardsAccountsLimitsRead,
	PermissionCreditCardsAccountsRead,
	PermissionCreditCardsAccountsTransactionsRead,
	PermissionCreditFixedIncomesRead,
	PermissionCustomersBusinessAdittionalInfoRead,
	PermissionCustomersBusinessIdentificationsRead,
	PermissionCustomersPersonalAdittionalInfoRead,
	PermissionCustomersPersonalIdentificationsRead,
	PermissionExchangesRead,
	PermissionFinancingsPaymentsRead,
	PermissionFinancingsRead,
	PermissionFinancingsScheduledInstalmentsRead,
	PermissionFinancingsWarrantiesRead,
	PermissionFundsRead,
	PermissionInvoiceFinancingsPaymentsRead,
	PermissionInvoiceFinancingsRead,
	PermissionInvoiceFinancingsScheduledInstalmentsRead,
	PermissionInvoiceFinancingsWarrantiesRead,
	PermissionLoansPaymentsRead,
	PermissionLoansRead,
	PermissionLoansScheduledInstalmentsRead,
	PermissionLoansWarrantiesRead,
	PermissionResourcesRead,
	PermissionTreasureTitlesRead,
	PermissionUnarrangedAccountsOverdraftPaymentsRead,
	PermissionUnarrangedAccountsOverdraftRead,
	PermissionUnarrangedAccountsOverdraftScheduledInstalmentsRead,
	PermissionUnarrangedAccountsOverdraftWarrantiesRead,
	PermissionVariableIncomesRead,
}

type PermissionGroup []Permission

var (
	// Dados Cadastrais PF.
	PermissionGroupPersonalRegistrationData = []Permission{
		PermissionCustomersPersonalIdentificationsRead,
		PermissionResourcesRead,
	}
	// Informações complementares PF.
	PermissionGroupPersonalAdditionalInfo PermissionGroup = []Permission{
		PermissionCustomersPersonalAdittionalInfoRead,
		PermissionResourcesRead,
	}
	// Dados Cadastrais PJ.
	PermissionGroupBusinessRegistrationData PermissionGroup = []Permission{
		PermissionCustomersBusinessIdentificationsRead,
		PermissionResourcesRead,
	}
	// Informações complementares PJ.
	PermissionGroupBusinessAdditionalInfo PermissionGroup = []Permission{
		PermissionCustomersBusinessAdittionalInfoRead,
		PermissionResourcesRead,
	}
	// Saldos.
	PermissionGroupBalances PermissionGroup = []Permission{
		PermissionAccountsRead,
		PermissionAccountsBalanceRead,
		PermissionResourcesRead,
	}
	// Limites.
	PermissionGroupLimits PermissionGroup = []Permission{
		PermissionAccountsRead,
		PermissionAccountsOverdraftLimitsRead,
		PermissionResourcesRead,
	}
	// Extratos.
	PermissionGroupStatements PermissionGroup = []Permission{
		PermissionAccountsRead,
		PermissionAccountsTransactionsRead,
		PermissionResourcesRead,
	}
	// Limites.
	PermissionGroupCreditCardLimits PermissionGroup = []Permission{
		PermissionCreditCardsAccountsRead,
		PermissionCreditCardsAccountsLimitsRead,
		PermissionResourcesRead,
	}
	// Transações.
	PermissionGroupCreditCardTransactions PermissionGroup = []Permission{
		PermissionCreditCardsAccountsRead,
		PermissionCreditCardsAccountsTransactionsRead,
		PermissionResourcesRead,
	}
	// Faturas.
	PermissionGroupCreditCardBills PermissionGroup = []Permission{
		PermissionCreditCardsAccountsRead,
		PermissionCreditCardsAccountsBillsRead,
		PermissionCreditCardsAccountsBillsTransactionsRead,
		PermissionResourcesRead,
	}
	// Bills.
	PermissionGroupContractData PermissionGroup = []Permission{
		PermissionLoansRead,
		PermissionLoansWarrantiesRead,
		PermissionLoansScheduledInstalmentsRead,
		PermissionLoansPaymentsRead,
		PermissionFinancingsRead,
		PermissionFinancingsWarrantiesRead,
		PermissionFinancingsScheduledInstalmentsRead,
		PermissionFinancingsPaymentsRead,
		PermissionUnarrangedAccountsOverdraftRead,
		PermissionUnarrangedAccountsOverdraftWarrantiesRead,
		PermissionUnarrangedAccountsOverdraftScheduledInstalmentsRead,
		PermissionUnarrangedAccountsOverdraftPaymentsRead,
		PermissionInvoiceFinancingsRead,
		PermissionInvoiceFinancingsWarrantiesRead,
		PermissionInvoiceFinancingsScheduledInstalmentsRead,
		PermissionInvoiceFinancingsPaymentsRead,
		PermissionResourcesRead,
	}
	// Dados da Operação.
	PermissionGroupInvestimentOperationalData PermissionGroup = []Permission{
		PermissionBankFixedIncomesRead,
		PermissionCreditFixedIncomesRead,
		PermissionFundsRead,
		PermissionVariableIncomesRead,
		PermissionTreasureTitlesRead,
		PermissionResourcesRead,
	}
	// Dados da Operação.
	PermissionGroupExchangeOperationalData PermissionGroup = []Permission{
		PermissionExchangesRead,
		PermissionResourcesRead,
	}
)

var PermissionGroups = []PermissionGroup{
	PermissionGroupPersonalRegistrationData,
	PermissionGroupPersonalAdditionalInfo,
	PermissionGroupBusinessRegistrationData,
	PermissionGroupBusinessAdditionalInfo,
	PermissionGroupBalances,
	PermissionGroupLimits,
	PermissionGroupStatements,
	PermissionGroupCreditCardLimits,
	PermissionGroupCreditCardTransactions,
	PermissionGroupCreditCardBills,
	PermissionGroupContractData,
	PermissionGroupInvestimentOperationalData,
	PermissionGroupExchangeOperationalData,
}

type RejectionInfo struct {
	RejectedBy RejectedBy
	Reason     RejectionReason
}

type RejectedBy string

const (
	RejectedByUser  RejectedBy = "USER"
	RejectedByASPSP RejectedBy = "ASPSP"
	RejectedByTPP   RejectedBy = "TPP"
)

type RejectionReason string

const (
	RejectionReasonConsentExpired           RejectionReason = "CONSENT_EXPIRED"
	RejectionReasonCustomerManuallyRejected RejectionReason = "CUSTOMER_MANUALLY_REJECTED"
	RejectionReasonCustomerManuallyRevoked  RejectionReason = "CUSTOMER_MANUALLY_REVOKED"
	RejectionReasonConsentMaxDateReached    RejectionReason = "CONSENT_MAX_DATE_REACHED"
	RejectionReasonConsentTechnicalIssue    RejectionReason = "CONSENT_TECHNICAL_ISSUE"
	RejectionReasonInternalSecurityReason   RejectionReason = "INTERNAL_SECURITY_REASON"
)

type Extension struct {
	ExpirationDateTime         *timex.DateTime `bson:"expires_at,omitempty"`
	PreviousExpirationDateTime *timex.DateTime `bson:"previous_expires_at,omitempty"`
	UserCPF                    string          `bson:"user_cpf"`
	BusinessCNPJ               string          `bson:"business_cnpj,omitempty"`
	RequestDateTime            timex.DateTime  `bson:"requested_at"`
	UserIPAddress              string          `bson:"user_ip,omitempty"`
	UserAgent                  string          `bson:"user_agent,omitempty"`
}

func consentID() string {
	return fmt.Sprintf("urn:mockbank:%s", uuid.NewString())
}
