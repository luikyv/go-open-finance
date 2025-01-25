package resource

type Resource struct {
	ID     string
	Type   Type
	Status Status
}

type Type string

const (
	TypeAccount                    Type = "ACCOUNT"
	TypeCreditCardAccount          Type = "CREDIT_CARD_ACCOUNT"
	TypeLoan                       Type = "LOAN"
	TypeFinancing                  Type = "FINANCING"
	TypeUnarrangedAccountOverdraft Type = "UNARRANGED_ACCOUNT_OVERDRAFT"
	TypeInvoiceFinancing           Type = "INVOICE_FINANCING"
	TypeBankFixedIncome            Type = "BANK_FIXED_INCOME"
	TypeCreditFixedIncome          Type = "CREDIT_FIXED_INCOME"
	TypeVariableIncome             Type = "VARIABLE_INCOME"
	TypeTreasureTitle              Type = "TREASURE_TITLE"
	TypeFund                       Type = "FUND"
	TypeExchange                   Type = "EXCHANGE"
)

type Status string

const (
	StatusAvailable             = "AVAILABLE"
	StatusUnavailable           = "UNAVAILABLE"
	StatusTemporarilyUnvailable = "TEMPORARILY_UNAVAILABLE"
	StatusPendingAuthorization  = "PENDING_AUTHORISATION"
)
