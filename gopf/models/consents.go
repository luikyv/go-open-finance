package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-opf/gopf/constants"
)

type ConsentStatus string

const (
	ConsentStatusAuthorised            ConsentStatus = "AUTHORISED"
	ConsentStatusAwaitingAuthorisation ConsentStatus = "AWAITING_AUTHORISATION"
	ConsentStatusRejected              ConsentStatus = "REJECTED"
)

type ConsentPermission string

const (
	PermissionAccountsBalanceRead                                 ConsentPermission = "ACCOUNTS_BALANCES_READ"
	PermissionAccountsOverdraftLimitsRead                         ConsentPermission = "ACCOUNTS_OVERDRAFT_LIMITS_READ"
	PermissionAccountsRead                                        ConsentPermission = "ACCOUNTS_READ"
	PermissionAccountsTransactionsRead                            ConsentPermission = "ACCOUNTS_TRANSACTIONS_READ"
	PermissionBankFixedIncomesRead                                ConsentPermission = "BANK_FIXED_INCOMES_READ"
	PermissionCreditCardsAccountsBillsRead                        ConsentPermission = "CREDIT_CARDS_ACCOUNTS_BILLS_READ"
	PermissionCreditCardsAccountsBillsTransactionsRead            ConsentPermission = "CREDIT_CARDS_ACCOUNTS_BILLS_TRANSACTIONS_READ"
	PermissionCreditCardsAccountsLimitsRead                       ConsentPermission = "CREDIT_CARDS_ACCOUNTS_LIMITS_READ"
	PermissionCreditCardsAccountsRead                             ConsentPermission = "CREDIT_CARDS_ACCOUNTS_READ"
	PermissionCreditCardsAccountsTransactionsRead                 ConsentPermission = "CREDIT_CARDS_ACCOUNTS_TRANSACTIONS_READ"
	PermissionCreditFixedIncomesRead                              ConsentPermission = "CREDIT_FIXED_INCOMES_READ"
	PermissionCustomersBusinessAdittionalInfoRead                 ConsentPermission = "CUSTOMERS_BUSINESS_ADITTIONALINFO_READ"
	PermissionCustomersBusinessIdentificationsRead                ConsentPermission = "CUSTOMERS_BUSINESS_IDENTIFICATIONS_READ"
	PermissionCustomersPersonalAdittionalInfoRead                 ConsentPermission = "CUSTOMERS_PERSONAL_ADITTIONALINFO_READ"
	PermissionCustomersPersonalIdentificationsRead                ConsentPermission = "CUSTOMERS_PERSONAL_IDENTIFICATIONS_READ"
	PermissionExchangesRead                                       ConsentPermission = "EXCHANGES_READ"
	PermissionFinancingsPaymentsRead                              ConsentPermission = "FINANCINGS_PAYMENTS_READ"
	PermissionFinancingsRead                                      ConsentPermission = "FINANCINGS_READ"
	PermissionFinancingsScheduledInstalmentsRead                  ConsentPermission = "FINANCINGS_SCHEDULED_INSTALMENTS_READ"
	PermissionFinancingsWarrantiesRead                            ConsentPermission = "FINANCINGS_WARRANTIES_READ"
	PermissionFundsRead                                           ConsentPermission = "FUNDS_READ"
	PermissionInvoiceFinancingsPaymentsRead                       ConsentPermission = "INVOICE_FINANCINGS_PAYMENTS_READ"
	PermissionInvoiceFinancingsRead                               ConsentPermission = "INVOICE_FINANCINGS_READ"
	PermissionInvoiceFinancingsScheduledInstalmentsRead           ConsentPermission = "INVOICE_FINANCINGS_SCHEDULED_INSTALMENTS_READ"
	PermissionInvoiceFinancingsWarrantiesRead                     ConsentPermission = "INVOICE_FINANCINGS_WARRANTIES_READ"
	PermissionLoansPaymentsRead                                   ConsentPermission = "LOANS_PAYMENTS_READ"
	PermissionLoansRead                                           ConsentPermission = "LOANS_READ"
	PermissionLoansScheduledInstalmentsRead                       ConsentPermission = "LOANS_SCHEDULED_INSTALMENTS_READ"
	PermissionLoansWarrantiesRead                                 ConsentPermission = "LOANS_WARRANTIES_READ"
	PermissionResourcesRead                                       ConsentPermission = "RESOURCES_READ"
	PermissionTreasureTitlesRead                                  ConsentPermission = "TREASURE_TITLES_READ"
	PermissionUnarrangedAccountsOverdraftPaymentsRead             ConsentPermission = "UNARRANGED_ACCOUNTS_OVERDRAFT_PAYMENTS_READ"
	PermissionUnarrangedAccountsOverdraftRead                     ConsentPermission = "UNARRANGED_ACCOUNTS_OVERDRAFT_READ"
	PermissionUnarrangedAccountsOverdraftScheduledInstalmentsRead ConsentPermission = "UNARRANGED_ACCOUNTS_OVERDRAFT_SCHEDULED_INSTALMENTS_READ"
	PermissionUnarrangedAccountsOverdraftWarrantiesRead           ConsentPermission = "UNARRANGED_ACCOUNTS_OVERDRAFT_WARRANTIES_READ"
	PermissionVariableIncomesRead                                 ConsentPermission = "VARIABLE_INCOMES_READ"
)

func (permission ConsentPermission) Scope() goidc.Scope {
	p := string(permission)
	switch {
	case strings.HasPrefix(p, "ACCOUNTS_"):
		return constants.ScopeAccounts
	case strings.HasPrefix(p, "BANK_FIXED_INCOMES_"):
		return constants.ScopeBankFixedIncomes
	case strings.HasPrefix(p, "CREDIT_CARDS_ACCOUNTS_"):
		return constants.ScopeCreditCardAccounts
	case strings.HasPrefix(p, "CREDIT_FIXED_INCOMES_"):
		return constants.ScopeCreditFixedIncomes
	case strings.HasPrefix(p, "CUSTOMERS_"):
		return constants.ScopeCustomers
	case strings.HasPrefix(p, "EXCHANGES_"):
		return constants.ScopeExchanges
	case strings.HasPrefix(p, "FINANCINGS_"):
		return constants.ScopeFinancings
	case strings.HasPrefix(p, "FUNDS_"):
		return constants.ScopeFunds
	case strings.HasPrefix(p, "INVOICE_FINANCINGS_"):
		return constants.ScopeInvoiceFinancings
	case strings.HasPrefix(p, "LOANS_"):
		return constants.ScopeLoans
	case strings.HasPrefix(p, "TREASURE_TITLES_"):
		return constants.ScopeTreasureTitles
	case strings.HasPrefix(p, "UNARRANGED_ACCOUNTS_"):
		return constants.ScopeUnarrangedAccountsOverdraft
	case strings.HasPrefix(p, "VARIABLE_INCOMES_"):
		return constants.ScopeVariableIncomes
	default:
		return constants.ScopeResources
	}
}

var ConsentPermissions = []ConsentPermission{
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

type PermissionGroup []ConsentPermission

var (
	// Dados Cadastrais PF.
	PermissionGroupPersonalRegistrationData = []ConsentPermission{
		PermissionCustomersPersonalIdentificationsRead,
		PermissionResourcesRead,
	}
	// Informações complementares PF.
	PermissionGroupPersonalAdditionalInfo PermissionGroup = []ConsentPermission{
		PermissionCustomersPersonalAdittionalInfoRead,
		PermissionResourcesRead,
	}
	// Dados Cadastrais PJ.
	PermissionGroupBusinessRegistrationData PermissionGroup = []ConsentPermission{
		PermissionCustomersBusinessIdentificationsRead,
		PermissionResourcesRead,
	}
	// Informações complementares PJ.
	PermissionGroupBusinessAdditionalInfo PermissionGroup = []ConsentPermission{
		PermissionCustomersBusinessAdittionalInfoRead,
		PermissionResourcesRead,
	}
	// Saldos.
	PermissionGroupBalances PermissionGroup = []ConsentPermission{
		PermissionAccountsRead,
		PermissionAccountsBalanceRead,
		PermissionResourcesRead,
	}
	// Limites.
	PermissionGroupLimits PermissionGroup = []ConsentPermission{
		PermissionAccountsRead,
		PermissionAccountsOverdraftLimitsRead,
		PermissionResourcesRead,
	}
	// Extratos.
	PermissionGroupStatements PermissionGroup = []ConsentPermission{
		PermissionAccountsRead,
		PermissionAccountsTransactionsRead,
		PermissionResourcesRead,
	}
	// Limites.
	PermissionGroupCreditCardLimits PermissionGroup = []ConsentPermission{
		PermissionCreditCardsAccountsRead,
		PermissionCreditCardsAccountsLimitsRead,
		PermissionResourcesRead,
	}
	// Transações.
	PermissionGroupCreditCardTransactions PermissionGroup = []ConsentPermission{
		PermissionCreditCardsAccountsRead,
		PermissionCreditCardsAccountsTransactionsRead,
		PermissionResourcesRead,
	}
	// Faturas.
	PermissionGroupCreditCardBills PermissionGroup = []ConsentPermission{
		PermissionCreditCardsAccountsRead,
		PermissionCreditCardsAccountsBillsRead,
		PermissionCreditCardsAccountsBillsTransactionsRead,
		PermissionResourcesRead,
	}
	// Bills.
	PermissionGroupContractData PermissionGroup = []ConsentPermission{
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
	PermissionGroupInvestimentOperationalData PermissionGroup = []ConsentPermission{
		PermissionBankFixedIncomesRead,
		PermissionCreditFixedIncomesRead,
		PermissionFundsRead,
		PermissionVariableIncomesRead,
		PermissionTreasureTitlesRead,
		PermissionResourcesRead,
	}
	// Dados da Operação.
	PermissionGroupExchangeOperationalData PermissionGroup = []ConsentPermission{
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

type Consent struct {
	ID                 string              `bson:"_id"`
	Status             ConsentStatus       `bson:"status"`
	UserCPF            string              `bson:"user_cpf"`
	BusinessCNPJ       string              `bson:"business_cnpj,omitempty"`
	ClientId           string              `bson:"client_id"`
	Permissions        []ConsentPermission `bson:"permissions"`
	CreatedAtTimestamp int64               `bson:"created_at"`
	UpdatedAtTimestamp int64               `bson:"updated_at"`
	ExpiresAtTimestamp int64               `bson:"expires_at,omitempty"`
}

func (c Consent) IsExpired() bool {
	now := time.Now().Unix()
	return (c.Status == ConsentStatusAwaitingAuthorisation && now-c.CreatedAtTimestamp > 3600) ||
		(c.Status == ConsentStatusAuthorised && now > c.ExpiresAtTimestamp)
}

type ConsentRequest struct {
	Data struct {
		LoggedUser         LoggedUser          `json:"loggedUser"`
		BusinessEntity     *BusinessEntity     `json:"businessEntity,omitempty"`
		Permissions        []ConsentPermission `json:"permissions"`
		ExpirationDateTime *DateTime           `json:"expirationDateTime,omitempty"`
	} `json:"data"`
}

type ConsentRequestV3 struct {
	ConsentRequest
}

func NewConsentV3(req ConsentRequestV3, clientID string) Consent {
	now := time.Now()
	consent := Consent{
		ID:                 consentID(),
		Status:             ConsentStatusAwaitingAuthorisation,
		UserCPF:            req.Data.LoggedUser.Document.Identification,
		ClientId:           clientID,
		Permissions:        req.Data.Permissions,
		CreatedAtTimestamp: now.Unix(),
		UpdatedAtTimestamp: now.Unix(),
	}

	if req.Data.BusinessEntity != nil {
		consent.BusinessCNPJ = req.Data.BusinessEntity.Document.Identification
	}

	if req.Data.ExpirationDateTime != nil {
		consent.ExpiresAtTimestamp = req.Data.ExpirationDateTime.Unix()
	}

	return consent
}

type ConsentResponse struct {
	ConsentID            string              `json:"consentId"`
	CreationDateTime     DateTime            `json:"creationDateTime"`
	Status               ConsentStatus       `json:"status"`
	StatusUpdateDateTime DateTime            `json:"statusUpdateDateTime"`
	Permissions          []ConsentPermission `json:"permissions"`
	ExpirationDateTime   *DateTime           `json:"expirationDateTime,omitempty"`
}

type ConsentResponseV3 struct {
	ConsentResponse
}

func (consent Consent) NewConsentResponseV3() Response {
	resp := ConsentResponseV3{}

	resp.ConsentID = consent.ID
	resp.CreationDateTime = DateTimeUnix(consent.CreatedAtTimestamp)
	resp.Status = consent.Status
	resp.StatusUpdateDateTime = DateTimeUnix(consent.CreatedAtTimestamp)
	resp.Permissions = consent.Permissions
	if consent.ExpiresAtTimestamp != 0 {
		expiresAt := DateTimeUnix(consent.ExpiresAtTimestamp)
		resp.ExpirationDateTime = &expiresAt
	}

	return Response{
		Data: resp,
		Meta: Meta{
			RequestDateTime: DateTimeNow(),
		},
		Links: Links{
			Self: fmt.Sprintf("%s/consents/%s", constants.BaseURLConsentsV3, consent.ID),
		},
	}
}

func consentID() string {
	return fmt.Sprintf("%s:%s", constants.Namespace, uuid.NewString())
}
