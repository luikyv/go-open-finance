package creditcard

import (
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/api/middleware"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/mock"
	"github.com/luikyv/go-open-finance/internal/page"
)

type APIRouterV2 struct {
	host           string
	service        Service
	consentService consent.Service
	op             *provider.Provider
}

func NewAPIRouterV2(host string, service Service, consentService consent.Service, op *provider.Provider) APIRouterV2 {
	return APIRouterV2{
		host:           host,
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (router APIRouterV2) Register(mux *http.ServeMux) {
	creditCardMux := http.NewServeMux()

	handler := router.getAccountsHandler()
	handler = consent.PermissionMiddlewareWithPagination(handler, router.consentService, consent.PermissionCreditCardsAccountsRead)
	handler = middleware.AuthScopesWithPagination(handler, router.op, goidc.ScopeOpenID, consent.ScopeID)
	handler = middleware.FAPIIDWithPagination(handler)
	creditCardMux.Handle("GET /open-banking/credit-cards-accounts/v2/accounts", handler)

	handler = router.getAccountHandler()
	handler = consent.PermissionMiddlewareWithPagination(handler, router.consentService, consent.PermissionCreditCardsAccountsRead)
	handler = middleware.AuthScopesWithPagination(handler, router.op, goidc.ScopeOpenID, consent.ScopeID)
	handler = middleware.FAPIIDWithPagination(handler)
	creditCardMux.Handle("GET /open-banking/credit-cards-accounts/v2/accounts/{id}", handler)

	handler = creditCardMux
	handler = middleware.Meta(handler, router.host)
	mux.Handle("/open-banking/credit-cards-accounts/v2/", handler)
}

func (router APIRouterV2) getAccountsHandler() http.Handler {
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
		api.WriteJSON(w, resp, http.StatusOK)
	})
}

func (router APIRouterV2) getAccountHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		cardID := r.PathValue("id")

		acc, err := router.service.account(r.Context(), cardID, consentID)
		if err != nil {
			writeErrorV2(w, err, true)
			return
		}

		resp := toAccountResponseV2(acc, reqURL)
		api.WriteJSON(w, resp, http.StatusOK)
	})
}

// func (router APIRouterV2) getAccountLimitsHandler() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
// 		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
// 		cardID := r.PathValue("id")

// 		acc, err := router.service.account(r.Context(), cardID, consentID)
// 		if err != nil {
// 			writeErrorV2(w, err, true)
// 			return
// 		}

// 		resp := toAccountLimitsResponseV2(acc, reqURL)
// 		api.WriteJSON(w, resp, http.StatusOK)
// 	})
// }

type accountsResponseV2 struct {
	Data  []accountV2 `json:"data"`
	Meta  api.Meta    `json:"meta"`
	Links api.Links   `json:"links"`
}

type accountV2 struct {
	ID          string  `json:"creditCardAccountId"`
	BrandName   string  `json:"brandName"`
	CompanyCNPJ string  `json:"companyCnpj"`
	Name        string  `json:"name"`
	Type        Type    `json:"productType"`
	Network     Network `json:"creditCardNetwork"`
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
			ID:          acc.ID,
			BrandName:   mock.MockBankBrand,
			CompanyCNPJ: mock.MockBankCNPJ,
			Name:        acc.Name,
			Type:        acc.Type,
			Network:     acc.Network,
		})
	}

	return resp
}

type accountResponseV2 struct {
	Data struct {
		Name          string            `json:"name"`
		Type          Type              `json:"productType"`
		Network       Network           `json:"creditCardNetwork"`
		PaymentMethod []paymentMethodV2 `json:"paymentMethod"`
	} `json:"data"`
	Meta  api.Meta  `json:"meta"`
	Links api.Links `json:"links"`
}

type paymentMethodV2 struct {
	CardNumber string `json:"identificationNumber"`
	IsMultiple bool   `json:"isMultipleCreditCard"`
}

func toAccountResponseV2(acc Account, reqURL string) accountResponseV2 {
	resp := accountResponseV2{
		Data: struct {
			Name          string            `json:"name"`
			Type          Type              `json:"productType"`
			Network       Network           `json:"creditCardNetwork"`
			PaymentMethod []paymentMethodV2 `json:"paymentMethod"`
		}{
			Name:    acc.Name,
			Type:    acc.Type,
			Network: acc.Network,
			PaymentMethod: []paymentMethodV2{
				{
					CardNumber: acc.MainCard.Number,
				},
			},
		},
		Meta: api.NewSingleRecordMeta(),
		Links: api.Links{
			Self: reqURL,
		},
	}

	for _, card := range acc.AdditionalCards {
		resp.Data.PaymentMethod = append(resp.Data.PaymentMethod, paymentMethodV2{
			CardNumber: card.Number,
		})
	}

	return resp
}

// type accountLimitsResponseV2 struct {
// 	Data  []accountLimitV2 `json:"data"`
// 	Meta  api.Meta         `json:"meta"`
// 	Links api.Links        `json:"links"`
// }

// type accountLimitV2 struct {
// 	CreditLineLimitType  string           `json:"creditLineLimitType"`
// 	ConsolidationType    string           `json:"consolidationType"`
// 	IdentificationNumber string           `json:"identificationNumber"`
// 	IsLimitFlexible      bool             `json:"isLimitFlexible"`
// 	LimitAmount          amountResponseV2 `json:"limitAmount"`
// 	UsedAmount           amountResponseV2 `json:"usedAmount"`
// 	AvailableAmount      amountResponseV2 `json:"availableAmount"`
// }

// type amountResponseV2 struct {
// 	Amount   string `json:"amount"`
// 	Currency string `json:"currency"`
// }

// func toAccountLimitsResponseV2(acc Account, reqURL string) accountLimitsResponseV2 {
// 	resp := accountLimitsResponseV2{
// 		Meta: api.NewSingleRecordMeta(),
// 		Links: api.Links{
// 			Self: reqURL,
// 		},
// 	}

// 	cards := append([]Card{acc.MainCard}, acc.AdditionalCards...)
// 	for _, card := range cards {
// 		resp.Data = append(resp.Data, accountLimitV2{
// 			CreditLineLimitType:  "LIMITE_CREDITO_TOTAL",
// 			ConsolidationType:    "CONSOLIDADO",
// 			IdentificationNumber: last4Digits(card.Number),
// 			LimitAmount: amountResponseV2{
// 				Amount:   acc.LimitAmount,
// 				Currency: defaultCurrency,
// 			},
// 			UsedAmount: amountResponseV2{
// 				Amount:   acc.UsedAmount,
// 				Currency: defaultCurrency,
// 			},
// 			AvailableAmount: amountResponseV2{
// 				Amount:   acc.LimitAmount,
// 				Currency: defaultCurrency,
// 			},
// 		})
// 	}

// 	return resp
// }

func writeErrorV2(w http.ResponseWriter, err error, _ bool) {
	api.WriteError(w, err)
}

// func last4Digits(cardNumber string) string {
// 	length := len(cardNumber)
// 	return cardNumber[length-4 : length-1]
// }
