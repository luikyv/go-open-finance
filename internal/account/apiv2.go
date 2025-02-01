package account

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/mock"
	"github.com/luikyv/go-open-finance/internal/page"
	"github.com/luikyv/go-open-finance/internal/timex"
)

const (
	dateTimeMillisFormat = "2006-01-02T15:04:05.000Z"
)

type APIHandlerV2 struct {
	service Service
}

func NewAPIHandlerV2(service Service) APIHandlerV2 {
	return APIHandlerV2{
		service: service,
	}
}

func (router APIHandlerV2) GetAccountsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		pag, err := api.NewPagination(r)
		if err != nil {
			writeErrorV2(w, api.NewError("INVALID_PARAMETER", http.StatusUnprocessableEntity, err.Error()), true)
			return
		}

		accs, err := router.service.accounts(r.Context(), consentID, pag)
		if err != nil {
			writeErrorV2(w, err, true)
			return
		}

		resp := toAccountsResponseV2(accs, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err, true)
			return
		}
	})
}

func (router APIHandlerV2) GetAccountHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		accID := r.PathValue("id")

		acc, err := router.service.account(r.Context(), accID, consentID)
		if err != nil {
			writeErrorV2(w, err, true)
			return
		}

		resp := toAccountResponseV2(acc, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err, true)
			return
		}
	})
}

func (router APIHandlerV2) GetAccountBalancesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		accID := r.PathValue("id")

		acc, err := router.service.account(r.Context(), accID, consentID)
		if err != nil {
			writeErrorV2(w, err, true)
			return
		}

		resp := toBalancesResponseV2(acc, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err, true)
			return
		}
	})
}

func (router APIHandlerV2) GetAccountTransactionsHandler() http.Handler {
	return router.getAccountTransactionsHandler(false, false)
}

func (router APIHandlerV2) GetAccountTransactionsCurrentHandler() http.Handler {
	return router.getAccountTransactionsHandler(true, true)
}

func (router APIHandlerV2) getAccountTransactionsHandler(current bool, pagination bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		accID := r.PathValue("id")
		pag, err := api.NewPagination(r)
		if err != nil {
			writeErrorV2(w, api.NewError("INVALID_PARAMETER", http.StatusUnprocessableEntity, err.Error()), pagination)
			return
		}

		filter, err := newTransactionFilter(r, current)
		if err != nil {
			writeErrorV2(w, err, pagination)
			return
		}

		trs, err := router.service.transactions(r.Context(), accID, consentID, pag, filter)
		if err != nil {
			writeErrorV2(w, err, pagination)
			return
		}

		resp := toAccountTransactionsResponseV2(trs, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err, pagination)
			return
		}
	})
}

func (router APIHandlerV2) GetAccountOverdraftLimitsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		accID := r.PathValue("id")

		acc, err := router.service.account(r.Context(), accID, consentID)
		if err != nil {
			writeErrorV2(w, err, true)
			return
		}

		resp := toOverdraftLimitsResponseV2(acc, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err, true)
			return
		}
	})
}

type accountsResponseV2 struct {
	Data  []accountV2 `json:"data"`
	Meta  api.Meta    `json:"meta"`
	Links api.Links   `json:"links"`
}

type accountV2 struct {
	BrandName   string `json:"brandName"`
	CompanyCNPJ string `json:"companyCnpj"`
	Type        Type   `json:"type"`
	CompeCode   string `json:"compeCode"`
	BranchCode  string `json:"branchCode"`
	Number      string `json:"number"`
	CheckDigit  string `json:"checkDigit"`
	AccountID   string `json:"accountId"`
}

func toAccountsResponseV2(accs page.Page[Account], reqURL string) accountsResponseV2 {
	resp := accountsResponseV2{
		Data: []accountV2{},
		Meta: api.NewPaginatedMeta(accs),
		Links: api.Links{
			Self: reqURL,
		},
	}
	for _, acc := range accs.Records {
		resp.Data = append(resp.Data, accountV2{
			BrandName:   mock.MockBankBrand,
			CompanyCNPJ: mock.MockBankCNPJ,
			Type:        acc.Type,
			CompeCode:   DefaultCompeCode,
			BranchCode:  DefaultBranch,
			Number:      acc.Number,
			CheckDigit:  DefaultCheckDigit,
			AccountID:   acc.ID,
		})
	}

	return resp
}

type accountResponseV2 struct {
	Data struct {
		Type       Type    `json:"type"`
		SubType    SubType `json:"subtype"`
		CompeCode  string  `json:"compeCode"`
		BranchCode string  `json:"branchCode"`
		Number     string  `json:"number"`
		CheckDigit string  `json:"checkDigit"`
		Currency   string  `json:"currency"`
	} `json:"data"`
	Meta  api.Meta  `json:"meta"`
	Links api.Links `json:"links"`
}

func toAccountResponseV2(acc Account, reqURL string) accountResponseV2 {
	return accountResponseV2{
		Data: struct {
			Type       Type    `json:"type"`
			SubType    SubType `json:"subtype"`
			CompeCode  string  `json:"compeCode"`
			BranchCode string  `json:"branchCode"`
			Number     string  `json:"number"`
			CheckDigit string  `json:"checkDigit"`
			Currency   string  `json:"currency"`
		}{
			Type:       acc.Type,
			SubType:    acc.SubType,
			CompeCode:  DefaultCompeCode,
			BranchCode: DefaultBranch,
			Number:     acc.Number,
			CheckDigit: DefaultCheckDigit,
			Currency:   DefaultCurrency,
		},
		Meta: api.NewSingleRecordMeta(),
		Links: api.Links{
			Self: reqURL,
		},
	}
}

type balancesResponseV2 struct {
	Data struct {
		AvailableAmount             amountResponseV2 `json:"availableAmount"`
		BlockedAmount               amountResponseV2 `json:"blockedAmount"`
		AutomaticallyInvestedAmount amountResponseV2 `json:"automaticallyInvestedAmount"`
		UpdateDateTime              timex.DateTime   `json:"updateDateTime"`
	} `json:"data"`
	Meta  api.Meta  `json:"meta"`
	Links api.Links `json:"links"`
}

func toBalancesResponseV2(acc Account, reqURL string) balancesResponseV2 {
	return balancesResponseV2{
		Data: struct {
			AvailableAmount             amountResponseV2 `json:"availableAmount"`
			BlockedAmount               amountResponseV2 `json:"blockedAmount"`
			AutomaticallyInvestedAmount amountResponseV2 `json:"automaticallyInvestedAmount"`
			UpdateDateTime              timex.DateTime   `json:"updateDateTime"`
		}{
			AvailableAmount: amountResponseV2{
				Amount:   acc.Balance.AvailableAmount,
				Currency: DefaultCurrency,
			},
			BlockedAmount: amountResponseV2{
				Amount:   acc.Balance.BlockedAmount,
				Currency: DefaultCurrency,
			},
			AutomaticallyInvestedAmount: amountResponseV2{
				Amount:   acc.Balance.AutomaticallyInvestedAmount,
				Currency: DefaultCurrency,
			},
			UpdateDateTime: timex.DateTimeNow(),
		},
		Meta:  api.NewSingleRecordMeta(),
		Links: api.NewLinks(reqURL),
	}
}

type transactionsResponseV2 struct {
	Data  []transactionResponseV2 `json:"data"`
	Meta  api.Meta                `json:"meta"`
	Links api.Links               `json:"links"`
}

type transactionResponseV2 struct {
	ID           string            `json:"transactionId"`
	Status       TransactionStatus `json:"completedAuthorisedPaymentType"`
	MovementType MovementType      `json:"creditDebitType"`
	Name         string            `json:"transactionName"`
	Type         TransactionType   `json:"type"`
	Amount       amountResponseV2  `json:"transactionAmount"`
	DateTime     string            `json:"transactionDateTime"`
}

func toAccountTransactionsResponseV2(trs page.Page[Transaction], reqURL string) transactionsResponseV2 {
	resp := transactionsResponseV2{
		Data:  []transactionResponseV2{},
		Meta:  api.NewMeta(),
		Links: api.NewPaginatedLinks(reqURL, trs),
	}

	for _, tr := range trs.Records {
		resp.Data = append(resp.Data, transactionResponseV2{
			ID:           tr.ID,
			Status:       tr.Status,
			MovementType: tr.MovementType,
			Name:         tr.Name,
			Type:         tr.Type,
			Amount: amountResponseV2{
				Amount:   tr.Amount,
				Currency: DefaultCurrency,
			},
			DateTime: tr.DateTime.Format(dateTimeMillisFormat),
		})
	}

	return resp
}

type overdraftLimitsResponseV2 struct {
	Data struct {
		Contracted *amountResponseV2 `json:"overdraftContractedLimit,omitempty"`
		Used       *amountResponseV2 `json:"overdraftUsedLimit,omitempty"`
		Unarranged *amountResponseV2 `json:"unarrangedOverdraftAmount,omitempty"`
	} `json:"data"`
	Meta  api.Meta  `json:"meta"`
	Links api.Links `json:"links"`
}

func toOverdraftLimitsResponseV2(acc Account, reqURL string) overdraftLimitsResponseV2 {
	resp := overdraftLimitsResponseV2{
		Meta:  api.NewSingleRecordMeta(),
		Links: api.NewLinks(reqURL),
	}

	if acc.OverdraftLimit.Contracted != "" {
		resp.Data.Contracted = &amountResponseV2{
			Amount:   acc.OverdraftLimit.Contracted,
			Currency: DefaultCurrency,
		}
	}

	if acc.OverdraftLimit.Used != "" {
		resp.Data.Used = &amountResponseV2{
			Amount:   acc.OverdraftLimit.Used,
			Currency: DefaultCurrency,
		}
	}

	if acc.OverdraftLimit.Unarranged != "" {
		resp.Data.Unarranged = &amountResponseV2{
			Amount:   acc.OverdraftLimit.Unarranged,
			Currency: DefaultCurrency,
		}
	}

	return resp
}

type amountResponseV2 struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func newTransactionFilter(r *http.Request, current bool) (transactionFilter, error) {
	now := timex.DateNow()
	filter := transactionFilter{
		from: now,
		to:   now,
	}

	from := r.URL.Query().Get("fromBookingDate")
	to := r.URL.Query().Get("toBookingDate")

	if from != "" {
		if to == "" {
			return transactionFilter{}, api.NewError("INVALID_PARAMETER",
				http.StatusUnprocessableEntity, "toBookingDate is required if fromBookingDate is informed")
		}

		fromDate, err := timex.ParseDate(from)
		if err != nil {
			return transactionFilter{}, api.NewError("INVALID_PARAMETER",
				http.StatusUnprocessableEntity, "invalid fromBookingDate")
		}
		filter.from = fromDate
	}

	if to != "" {
		if from == "" {
			return transactionFilter{}, api.NewError("INVALID_PARAMETER",
				http.StatusUnprocessableEntity, "fromBookingDate is required if toBookingDate is informed")
		}

		toDate, err := timex.ParseDate(to)
		if err != nil {
			return transactionFilter{}, api.NewError("INVALID_PARAMETER",
				http.StatusUnprocessableEntity, "invalid toBookingDate")
		}
		filter.to = toDate
	}

	if current {
		nowMinus7Days := now.AddDate(0, 0, -7)
		if filter.from.Before(nowMinus7Days) {
			return transactionFilter{}, api.NewError("INVALID_PARAMETER",
				http.StatusUnprocessableEntity, "fromBookingDate too far in the past")
		}

		if filter.to.Before(nowMinus7Days) {
			return transactionFilter{}, api.NewError("INVALID_PARAMETER",
				http.StatusUnprocessableEntity, "toBookingDate too far in the past")
		}
	}

	return filter, nil
}

func writeErrorV2(w http.ResponseWriter, err error, pagination bool) {
	if errors.Is(err, errAccountNotAllowed) {
		err := api.NewError("FORBIDDEN", http.StatusForbidden, errAccountNotAllowed.Error())
		if pagination {
			err = err.WithPagination()
		}
		api.WriteError(w, err)
		return
	}

	if errors.Is(err, errJointAccountPendingAuthorization) {
		err := api.NewError("STATUS_RESOURCE_AWAITING_AUTHORIZATION", http.StatusForbidden, errAccountNotAllowed.Error())
		if pagination {
			err = err.WithPagination()
		}
		api.WriteError(w, err)
		return
	}

	api.WriteError(w, err)
}
