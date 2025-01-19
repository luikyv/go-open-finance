package consent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/timex"
)

var (
	errBadRequest = api.NewError("INVALID_REQUEST", http.StatusBadRequest, "invalid request")
)

type APIHandlerV3 struct {
	host    string
	service Service
}

func NewAPIHandlerV3(host string, service Service) APIHandlerV3 {
	return APIHandlerV3{
		host:    host,
		service: service,
	}
}

func (router APIHandlerV3) CreateHandler() http.Handler {
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			api.WriteError(w, errBadRequest)
			return
		}
	})
}

func (router APIHandlerV3) GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("consent_id")
		c, err := router.service.Fetch(r.Context(), id)
		if err != nil {
			writeErrorV3(w, err)
			return
		}

		resp := toResponseV3(c, router.host)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			api.WriteError(w, errBadRequest)
			return
		}
	})
}

func (router APIHandlerV3) DeleteHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("consent_id")
		err := router.service.delete(r.Context(), id)
		if err != nil {
			writeErrorV3(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

type createRequestV3 struct {
	Data struct {
		LoggerUser struct {
			Document struct {
				Identification string `json:"identification"`
				Relation       string `json:"rel"`
			} `json:"document"`
		} `json:"loggedUser"`
		BusinessEntity *struct {
			Document struct {
				Identification string `json:"identification"`
				Relation       string `json:"rel"`
			} `json:"document"`
		} `json:"businessEntity,omitempty"`
		Permissions        []Permission
		ExpirationDateTime *timex.DateTime
	} `json:"data"`
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
		Links: api.Links{
			Self: host + "/open-banking/consents/v3/consents/" + c.ID,
		},
		Meta: api.NewMeta(),
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

func writeErrorV3(w http.ResponseWriter, err error) {
	if errors.Is(err, errAccessNotAllowed) {
		api.WriteError(w, api.NewError("FORBIDDEN", http.StatusForbidden, errAccessNotAllowed.Error()))
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

	api.WriteError(w, errBadRequest)
}
