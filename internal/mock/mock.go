package mock

import (
	"time"

	"github.com/luikyv/go-open-finance/internal/timex"
)

const (
	CPFWithJointAccount string = "96362357086"
	MockBankBrand       string = "MockBank"
	MockBankCNPJ        string = "58540569000120"
)

func IsJointAccountPendingAuth(consentAuthorizedAt timex.DateTime) bool {
	return timex.Now().Before(consentAuthorizedAt.Time.Add(30 * time.Second))
}
