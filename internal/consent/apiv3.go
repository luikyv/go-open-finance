package consent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/api/middleware"
	"github.com/luikyv/go-open-finance/internal/page"
	"github.com/luikyv/go-open-finance/internal/timex"
)

var (
	errBadRequest = api.NewError("INVALID_REQUEST", http.StatusBadRequest, "invalid request")
)

type APIRouterV3 struct {
	host    string
	service Service
	op      *provider.Provider
}

func NewAPIRouterV3(host string, service Service, op *provider.Provider) APIRouterV3 {
	return APIRouterV3{
		host:    host,
		service: service,
		op:      op,
	}
}

func (router APIRouterV3) Register(mux *http.ServeMux) {
	consentMux := http.NewServeMux()

	handler := router.CreateHandler()
	handler = middleware.AuthScopes(handler, router.op, Scope)
	consentMux.Handle("POST /open-banking/consents/v3/consents", handler)

	handler = router.GetHandler()
	handler = middleware.AuthScopes(handler, router.op, Scope)
	consentMux.Handle("GET /open-banking/consents/v3/consents/{id}", handler)

	handler = router.DeleteHandler()
	handler = middleware.AuthScopes(handler, router.op, Scope)
	consentMux.Handle("DELETE /open-banking/consents/v3/consents/{id}", handler)

	handler = router.ExtendHandler()
	handler = middleware.AuthScopes(handler, router.op, goidc.ScopeOpenID, ScopeID)
	consentMux.Handle("POST /open-banking/consents/v3/consents/{id}/extends", handler)

	handler = router.GetExtensionsHandler()
	handler = middleware.AuthScopes(handler, router.op, Scope)
	consentMux.Handle("GET /open-banking/consents/v3/consents/{id}/extensions", handler)

	handler = consentMux
	handler = middleware.FAPIID(handler)
	handler = middleware.Meta(handler, router.host)
	mux.Handle("/open-banking/consents/", handler)
}

func (router APIRouterV3) CreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req createRequestV3
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			api.WriteError(w, errBadRequest)
			return
		}

		if err := req.validate(); err != nil {
			writeErrorV3(w, err)
			return
		}

		consent := req.toConsent(r.Context())
		if err := router.service.create(r.Context(), consent); err != nil {
			writeErrorV3(w, err)
			return
		}

		resp := toResponseV3(consent, router.host)
		api.WriteJSON(w, resp, http.StatusCreated)
	})
}

func (router APIRouterV3) GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		c, err := router.service.Consent(r.Context(), id)
		if err != nil {
			writeErrorV3(w, err)
			return
		}

		resp := toResponseV3(c, router.host)
		api.WriteJSON(w, resp, http.StatusOK)
	})
}

func (router APIRouterV3) DeleteHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		err := router.service.delete(r.Context(), id)
		if err != nil {
			writeErrorV3(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func (router APIRouterV3) ExtendHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id != r.Context().Value(api.CtxKeyConsentID) {
			api.WriteError(w, errBadRequest)
			return
		}

		ip := r.Header.Get(headerCustomerIPAddress)
		if ip == "" {
			api.WriteError(w, errBadRequest)
			return
		}

		userAgent := r.Header.Get(headerCustomerUserAgent)
		if userAgent == "" {
			api.WriteError(w, errBadRequest)
			return
		}

		var req extendRequestV3
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			api.WriteError(w, errBadRequest)
			return
		}

		if err := req.validate(); err != nil {
			writeErrorV3(w, err)
			return
		}

		c, err := router.service.extend(r.Context(), id, req.toExtension(ip, userAgent))
		if err != nil {
			writeErrorV3(w, err)
			return
		}

		resp := toResponseV3(c, router.host)
		api.WriteJSON(w, resp, http.StatusCreated)
	})
}

func (router APIRouterV3) GetExtensionsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		pag, err := api.NewPagination(r)
		if err != nil {
			writeErrorV3(w, api.NewError("INVALID_PARAMETER", http.StatusUnprocessableEntity, err.Error()))
			return
		}

		exts, err := router.service.extensions(r.Context(), id, pag)
		if err != nil {
			writeErrorV3(w, err)
			return
		}

		resp := toExtensionsResponseV3(exts, router.host)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			api.WriteError(w, errBadRequest)
			return
		}
	})
}

type createRequestV3 struct {
	Data struct {
		LoggerUser         entityV3  `json:"loggedUser"`
		BusinessEntity     *entityV3 `json:"businessEntity,omitempty"`
		Permissions        []Permission
		ExpirationDateTime *timex.DateTime
	} `json:"data"`
}

type entityV3 struct {
	Document documentV3 `json:"document"`
}

type documentV3 struct {
	Identification string `json:"identification"`
	Relation       string `json:"rel"`
}

func (req createRequestV3) validate() error {
	for _, p := range req.Data.Permissions {
		if !slices.Contains(Permissions, p) {
			return api.NewError("INVALID_PERMISSION", http.StatusBadRequest, "invalid request")
		}
	}
	return nil
}

func (req createRequestV3) toConsent(ctx context.Context) Consent {
	now := timex.DateTimeNow()
	consent := Consent{
		ID:                   consentID(),
		Status:               StatusAwaitingAuthorization,
		UserCPF:              req.Data.LoggerUser.Document.Identification,
		Permissions:          req.Data.Permissions,
		CreationDateTime:     now,
		StatusUpdateDateTime: now,
		ExpirationDateTime:   req.Data.ExpirationDateTime,
		ClientID:             ctx.Value(api.CtxKeyClientID).(string),
	}

	if req.Data.BusinessEntity != nil {
		consent.BusinessCNPJ = req.Data.BusinessEntity.Document.Identification
	}

	return consent
}

type responseV3 struct {
	Data struct {
		ID                   string          `json:"consentId"`
		Status               Status          `json:"status"`
		Permissions          []Permission    `json:"permissions"`
		CreationDateTime     timex.DateTime  `json:"creationDateTime"`
		StatusUpdateDateTime timex.DateTime  `json:"statusUpdateDateTime"`
		ExpirationDateTime   *timex.DateTime `json:"expirationDateTime,omitempty"`
		Rejection            *struct {
			RejectedBy RejectedBy `json:"rejectedBy"`
			Reason     struct {
				Code RejectionReason `json:"code"`
			} `json:"reason"`
		} `json:"rejection,omitempty"`
	} `json:"data"`
	Links api.Links `json:"links"`
	Meta  api.Meta  `json:"meta"`
}

func toResponseV3(c Consent, host string) responseV3 {
	resp := responseV3{
		Links: api.NewLinks(host + "/open-banking/consents/v3/consents/" + c.ID),
		Meta:  api.NewMeta(),
	}
	resp.Data.ID = c.ID
	resp.Data.Status = c.Status
	resp.Data.Permissions = c.Permissions
	resp.Data.CreationDateTime = c.CreationDateTime
	resp.Data.StatusUpdateDateTime = c.StatusUpdateDateTime
	resp.Data.ExpirationDateTime = c.ExpirationDateTime

	if c.RejectionInfo != nil {
		resp.Data.Rejection = &struct {
			RejectedBy RejectedBy `json:"rejectedBy"`
			Reason     struct {
				Code RejectionReason `json:"code"`
			} `json:"reason"`
		}{
			RejectedBy: c.RejectionInfo.RejectedBy,
			Reason: struct {
				Code RejectionReason `json:"code"`
			}{
				Code: c.RejectionInfo.Reason,
			},
		}
	}

	return resp
}

type extendRequestV3 struct {
	Data struct {
		ExpirationDateTime *timex.DateTime
		LoggerUser         entityV3  `json:"loggedUser"`
		BusinessEntity     *entityV3 `json:"businessEntity,omitempty"`
	} `json:"data"`
}

func (r extendRequestV3) validate() error {
	return nil
}

func (r extendRequestV3) toExtension(ip, userAgent string) Extension {
	ext := Extension{
		ExpirationDateTime: r.Data.ExpirationDateTime,
		UserCPF:            r.Data.LoggerUser.Document.Identification,
		RequestDateTime:    timex.DateTimeNow(),
		UserIPAddress:      ip,
		UserAgent:          userAgent,
	}
	if r.Data.BusinessEntity != nil {
		ext.BusinessCNPJ = r.Data.BusinessEntity.Document.Identification
	}

	return ext
}

type extensionsResponseV3 struct {
	Data  []extensionResponseV3 `json:"data"`
	Links api.Links             `json:"links"`
	Meta  api.Meta              `json:"meta"`
}

type extensionResponseV3 struct {
	ExpirationDateTime         *timex.DateTime `json:"expirationDateTime,omitempty"`
	PreviousExpirationDateTime *timex.DateTime `json:"previousExpirationDateTime,omitempty"`
	LoggerUser                 entityV3        `json:"loggedUser"`
	RequestDateTime            timex.DateTime  `json:"requestDateTime"`
	CustomerIPAddress          string          `json:"xFapiCustomerIpAddress"`
	CustomerUserAgent          string          `json:"xCustomerUserAgent"`
}

func toExtensionsResponseV3(exts page.Page[Extension], reqURL string) extensionsResponseV3 {
	resp := extensionsResponseV3{
		Links: api.Links{
			Self: reqURL,
		},
		Meta: api.NewPaginatedMeta(exts),
	}

	for _, ext := range exts.Records {
		resp.Data = append(resp.Data, extensionResponseV3{
			LoggerUser: entityV3{
				Document: documentV3{
					Identification: ext.UserCPF,
					Relation:       defaultUserDocumentRelation,
				},
			},
			ExpirationDateTime:         ext.ExpirationDateTime,
			PreviousExpirationDateTime: ext.PreviousExpirationDateTime,
			RequestDateTime:            ext.RequestDateTime,
			CustomerIPAddress:          ext.UserIPAddress,
			CustomerUserAgent:          ext.UserAgent,
		})
	}

	return resp
}

func writeErrorV3(w http.ResponseWriter, err error) {
	if errors.Is(err, errAccessNotAllowed) {
		api.WriteError(w, api.NewError("FORBIDDEN", http.StatusForbidden, errAccessNotAllowed.Error()))
		return
	}

	if errors.Is(err, errExtensionNotAllowed) {
		api.WriteError(w, api.NewError("FORBIDDEN", http.StatusForbidden, errExtensionNotAllowed.Error()))
		return
	}

	if errors.Is(err, errInvalidPermissionGroup) {
		api.WriteError(w, api.NewError("COMBINACAO_PERMISSOES_INCORRETA", http.StatusUnprocessableEntity, errInvalidPermissionGroup.Error()))
		return
	}

	if errors.Is(err, errPersonalAndBusinessPermissionsTogether) {
		api.WriteError(w, api.NewError("PERMISSAO_PF_PJ_EM_CONJUNTO", http.StatusUnprocessableEntity, errPersonalAndBusinessPermissionsTogether.Error()))
		return
	}

	if errors.Is(err, errInvalidExpiration) {
		api.WriteError(w, api.NewError("DATA_EXPIRACAO_INVALIDA", http.StatusUnprocessableEntity, errInvalidExpiration.Error()))
		return
	}

	if errors.Is(err, errAlreadyRejected) {
		api.WriteError(w, api.NewError("CONSENTIMENTO_EM_STATUS_REJEITADO", http.StatusUnprocessableEntity, errAlreadyRejected.Error()))
		return
	}

	if errors.Is(err, errCannotExtendConsentNotAuthorized) {
		api.WriteError(w, api.NewError("ESTADO_CONSENTIMENTO_INVALIDO", http.StatusUnprocessableEntity, errCannotExtendConsentNotAuthorized.Error()))
		return
	}

	if errors.Is(err, errCannotExtendConsentForJointAccount) {
		api.WriteError(w, api.NewError("DEPENDE_MULTIPLA_ALCADA", http.StatusUnprocessableEntity, errCannotExtendConsentForJointAccount.Error()))
		return
	}

	api.WriteError(w, errBadRequest)
}
