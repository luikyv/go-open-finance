package creditcard

import (
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-open-finance/internal/timex"
)

var (
	Scope = goidc.NewScope("credit-cards-accounts")
)

const (
// defaultCurrency string = "BRL"
)

type Account struct {
	ID              string
	UserID          string
	Name            string
	Type            Type
	Network         Network
	MainCard        Card
	AdditionalCards []Card
	LimitAmount     string
	UsedAmount      string
	Biils           []Bill
}

type Card struct {
	Number string
}

type Type string

const (
	TypeClassicNational          Type = "CLASSIC_NACIONAL"
	TypeClassicInternational     Type = "CLASSIC_INTERNACIONAL"
	TypeGold                     Type = "GOLD"
	TypePlatinum                 Type = "PLATINUM"
	TypeInfinite                 Type = "INFINITE"
	TypeElectron                 Type = "ELECTRON"
	TypeStandardNational         Type = "STANDARD_NACIONAL"
	TypeStandardInternational    Type = "STANDARD_INTERNACIONAL"
	TypeElectronic               Type = "ELETRONIC"
	TypeBlack                    Type = "BLACK"
	TypeRedeshop                 Type = "REDESHOP"
	TypeMaestroMastercardMaestro Type = "MAESTRO_MASTERCARD_MAESTRO"
	TypeGreen                    Type = "GREEN"
	TypeBlue                     Type = "BLUE"
	TypeBluebox                  Type = "BLUEBOX"
	TypeLiberalProfessional      Type = "PROFISSIONAL_LIBERAL"
	TypeElectronicCheck          Type = "CHEQUE_ELETRONICO"
	TypeCorporate                Type = "CORPORATIVO"
	TypeBusiness                 Type = "EMPRESARIAL"
	TypePurchases                Type = "COMPRAS"
	TypeBasicNational            Type = "BASICO_NACIONAL"
	TypeBasicInternational       Type = "BASICO_INTERNACIONAL"
	TypeNanquim                  Type = "NANQUIM"
	TypeGraphite                 Type = "GRAFITE"
	TypeMore                     Type = "MAIS"
	TypeOthers                   Type = "OUTROS"
)

type Network string

const (
	NetworkVisa            Network = "VISA"
	NetworkMastercard      Network = "MASTERCARD"
	NetworkAmericanExpress Network = "AMERICAN_EXPRESS"
	NetworkDinersClub      Network = "DINERS_CLUB"
	NetworkHipercard       Network = "HIPERCARD"
	NetworkPrivateLabel    Network = "BANDEIRA_PROPRIA"
	NetworkElectronicCheck Network = "CHEQUE_ELETRONICO"
	NetworkElo             Network = "ELO"
	NetworkOthers          Network = "OUTRAS"
)

type Bill struct {
	ID          string
	DueDate     timex.Date
	TotalAmount string
}

type Payment struct {
	Date   timex.Date
	Mode   PaymentMode
	Amount string
}

type PaymentMode string

const (
	PaymentModeDebitCurrentAccount PaymentMode = "DEBITO_CONTA_CORRENTE"
	PaymentModeBankSlip            PaymentMode = "BOLETO_BANCARIO"
	PaymentModePayrollDeduction    PaymentMode = "AVERBACAO_FOLHA"
	PaymentModePix                 PaymentMode = "PIX"
)
