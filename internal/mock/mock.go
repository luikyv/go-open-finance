package mock

import (
	"time"

	"github.com/luikyv/go-open-finance/internal/timex"
)

const (
	CPFWithJointAccount string = "96362357086"
)

func IsJointAccountPendingAuth(consentCreateAt timex.DateTime) bool {
	return timex.Now().Before(consentCreateAt.Time.Add(3 * time.Minute))
}
