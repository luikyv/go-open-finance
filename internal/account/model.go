package account

import (
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-open-finance/internal/timex"
)

const (
	DefaultCompeCode  string = "001"
	DefaultBranch     string = "0001"
	DefaultCheckDigit string = "1"
	DefaultCurrency   string = "BRL"
)

var (
	Scope = goidc.NewScope("accounts")
)

type Account struct {
	ID             string
	UserID         string
	Number         string
	Type           Type
	SubType        SubType
	Balance        Balance
	Transactions   []Transaction
	OverdraftLimit OverdraftLimit
}

type Type string

const (
	TypeCheckingAccount Type = "CONTA_DEPOSITO_A_VISTA"
	TypeSavingsAccount  Type = "CONTA_POUPANCA"
	TypePrepaidPayment  Type = "CONTA_PAGAMENTO_PRE_PAGA"
)

type SubType string

const (
	SubTypeIndividual    SubType = "INDIVIDUAL"
	SubTypeJointSimple   SubType = "CONJUNTA_SIMPLES"
	SubTypeJointSolidary SubType = "CONJUNTA_SOLIDARIA"
)

type Balance struct {
	AvailableAmount             string
	BlockedAmount               string
	AutomaticallyInvestedAmount string
}

type Transaction struct {
	ID           string
	Status       TransactionStatus
	MovementType MovementType
	Name         string
	Type         TransactionType
	Amount       string
	DateTime     timex.DateTime
}

type TransactionStatus string

const (
	TransactionStatusCompleted   TransactionStatus = "TRANSACAO_EFETIVADA"
	TransactionStatusFutureEntry TransactionStatus = "LANCAMENTO_FUTURO"
	TransactionStatusProcessing  TransactionStatus = "TRANSACAO_PROCESSANDO"
)

type TransactionType string

const (
	TransactionTypeTed                       TransactionType = "TED"
	TransactionTypeDoc                       TransactionType = "DOC"
	TransactionTypePix                       TransactionType = "PIX"
	TransactionTypeTransferSameInstitution   TransactionType = "TRANSFERENCIAMESMAINSTITUICAO"
	TransactionTypeBoleto                    TransactionType = "BOLETO"
	TransactionTypeAgreementCollection       TransactionType = "CONVENIOARRECADACAO"
	TransactionTypeServicePackageFee         TransactionType = "PACOTETARIFASERVICOS"
	TransactionTypeSingleServiceFee          TransactionType = "TARIFASERVICOSAVULSOS"
	TransactionTypePayroll                   TransactionType = "FOLHAPAGAMENTO"
	TransactionTypeDeposit                   TransactionType = "DEPOSITO"
	TransactionTypeWithdrawal                TransactionType = "SAQUE"
	TransactionTypeCard                      TransactionType = "CARTAO"
	TransactionTypeOverdraftInterestCharges  TransactionType = "ENCARGOSJUROSCHEQUEESPECIAL"
	TransactionTypeFinancialInvestmentIncome TransactionType = "RENDIMENTOAPLICFINANCEIRA"
	TransactionTypeSalaryPortability         TransactionType = "PORTABILIDADESALARIO"
	TransactionTypeFinancialInvestmentRescue TransactionType = "RESGATEAPLICFINANCEIRA"
	TransactionTypeCreditOperation           TransactionType = "OPERACAOCREDITO"
	TransactionTypeOthers                    TransactionType = "OUTROS"
)

type MovementType string

const (
	MovementTypeCredit MovementType = "CREDITO"
	MovementTypeDebit  MovementType = "DEBITO"
)

type OverdraftLimit struct {
	Contracted string
	Used       string
	Unarranged string
}

type transactionFilter struct {
	from timex.Date
	to   timex.Date
}
