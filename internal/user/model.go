package user

import (
	"slices"

	"github.com/luikyv/go-open-finance/internal/timex"
)

type User struct {
	UserName      string
	Email         string
	CPF           string
	Name          string
	AccountNumber string
	CompanyCNPJs  []string
}

func (u User) OwnsCompany(cnpj string) bool {
	return slices.Contains(u.CompanyCNPJs, cnpj)
}

type Company struct {
	Name string
	CNPJ string
}

type PersonalIdentification struct {
	ID             string
	BrandName      string
	CivilName      string
	SocialName     string
	BirthDate      timex.Date
	MaritalStatus  MaritalStatus
	Sex            Sex
	CompanyCNPJ    string
	CPF            string
	Addresses      []Address
	Phones         []Phone
	Emails         []Email
	UpdateDateTime timex.DateTime
}

type MaritalStatus string

const (
	MaritalStatusSOLTEIRO               MaritalStatus = "SOLTEIRO"
	MaritalStatusCASADO                 MaritalStatus = "CASADO"
	MaritalStatusVIUVO                  MaritalStatus = "VIUVO"
	MaritalStatusSEPARADO_JUDICIALMENTE MaritalStatus = "SEPARADO_JUDICIALMENTE"
	MaritalStatusDIVORCIADO             MaritalStatus = "DIVORCIADO"
	MaritalStatusUNIAO_ESTAVEL          MaritalStatus = "UNIAO_ESTAVEL"
	MaritalStatusOUTRO                  MaritalStatus = "OUTRO"
)

type Sex string

const (
	SexMale   Sex = "MASCULINO"
	SexFemale Sex = "FEMININO"
	SexOther  Sex = "OTHER"
)

type Address struct {
	IsMain   bool
	Address  string
	TownName string
	PostCode string
	Country  string
}

type Phone struct {
	IsMain   bool
	Type     PhoneType
	AreaCode string
	Number   string
}

type PhoneType string

const (
	PhoneTypeLandline PhoneType = "FIXO"
	PhoneTypeMobile   PhoneType = "MOVEL"
	PhoneTypeOther    PhoneType = "OUTRO"
)

type Email struct {
	IsMain bool
	Email  string
}

type PersonalQualifications struct {
	CompanyCNPJ           string
	Occupation            Occupation
	OccupationDescription string
	UpdateDateTime        timex.DateTime
}

type Occupation string

const (
	OccupationRECEITA_FEDERAL Occupation = "RECEITA_FEDERAL"
	OccupationCBO             Occupation = "CBO"
	OccupationOUTRO           Occupation = "OUTRO"
)

type PersonalFinancialRelations struct {
	ProductServiceTypes          []ProductServiceType
	ProductServiceAdditionalInfo string
	Accounts                     []Account
	StartDateTime                timex.DateTime
	UpdateDateTime               timex.DateTime
}

type ProductServiceType string

const (
	ProductServiceTypeCONTA_DEPOSITO_A_VISTA   ProductServiceType = "CONTA_DEPOSITO_A_VISTA"
	ProductServiceTypeCONTA_POUPANCA           ProductServiceType = "CONTA_POUPANCA"
	ProductServiceTypeCONTA_PAGAMENTO_PRE_PAGA ProductServiceType = "CONTA_PAGAMENTO_PRE_PAGA"
	ProductServiceTypeCARTAO_CREDITO           ProductServiceType = "CARTAO_CREDITO"
	ProductServiceTypeOPERACAO_CREDITO         ProductServiceType = "OPERACAO_CREDITO"
	ProductServiceTypeSEGURO                   ProductServiceType = "SEGURO"
	ProductServiceTypePREVIDENCIA              ProductServiceType = "PREVIDENCIA"
	ProductServiceTypeINVESTIMENTO             ProductServiceType = "INVESTIMENTO"
	ProductServiceTypeOPERACOES_CAMBIO         ProductServiceType = "OPERACOES_CAMBIO"
	ProductServiceTypeCONTA_SALARIO            ProductServiceType = "CONTA_SALARIO"
	ProductServiceTypeCREDENCIAMENTO           ProductServiceType = "CREDENCIAMENTO"
	ProductServiceTypeOUTROS                   ProductServiceType = "OUTROS"
)

type Account struct {
	CompeCode  string
	Branch     string
	Number     string
	CheckDigit string
	Type       AccountType
	SubType    AccountSubType
}

type AccountType string

const (
	AccountTypeCONTA_DEPOSITO_A_VISTA   AccountType = "CONTA_DEPOSITO_A_VISTA"
	AccountTypeCONTA_POUPANCA           AccountType = "CONTA_POUPANCA"
	AccountTypeCONTA_PAGAMENTO_PRE_PAGA AccountType = "CONTA_PAGAMENTO_PRE_PAGA"
)

type AccountSubType string

const (
	AccountSubTypeINDIVIDUAL         AccountSubType = "INDIVIDUAL"
	AccountSubTypeCONJUNTA_SIMPLES   AccountSubType = "CONJUNTA_SIMPLES"
	AccountSubTypeCONJUNTA_SOLIDARIA AccountSubType = "CONJUNTA_SOLIDARIA"
)
