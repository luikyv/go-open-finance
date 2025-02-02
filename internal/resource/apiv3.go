package resource

import (
	"encoding/json"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/provider"
	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/api/middleware"
	"github.com/luikyv/go-open-finance/internal/consent"
	"github.com/luikyv/go-open-finance/internal/page"
)

type APIRouterV3 struct {
	host           string
	service        Service
	consentService consent.Service
	op             provider.Provider
}

func NewAPIRouterV3(host string, service Service, consentService consent.Service, op provider.Provider) APIRouterV3 {
	return APIRouterV3{
		host:           host,
		service:        service,
		consentService: consentService,
		op:             op,
	}
}

func (router APIRouterV3) Register(mux *http.ServeMux) {
	handler := router.getHandler()
	handler = consent.PermissionMiddleware(handler, router.consentService, consent.PermissionResourcesRead)
	handler = middleware.AuthScopes(handler, router.op, consent.ScopeID)
	handler = middleware.FAPIID(handler)
	handler = middleware.Meta(handler, router.host)
	mux.Handle("GET /open-banking/resources/v3/resources", handler)
}

func (router APIRouterV3) getHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		pag, err := api.NewPagination(r)
		if err != nil {
			writeErrorV3(w, api.NewError("INVALID_PARAMETER", http.StatusUnprocessableEntity, err.Error()))
			return
		}

		rs, err := router.service.resources(r.Context(), consentID, pag)
		if err != nil {
			writeErrorV3(w, err)
			return
		}

		resp := toResponseV3(rs, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			api.WriteError(w, err)
			return
		}
	})
}

type responseV3 struct {
	Data  []resourceV3 `json:"data"`
	Meta  api.Meta     `json:"meta"`
	Links api.Links    `json:"links"`
}

type resourceV3 struct {
	ID     string `json:"resourceId"`
	Type   Type   `json:"type"`
	Status Status `json:"status"`
}

func toResponseV3(rs page.Page[Resource], reqURL string) responseV3 {
	resp := responseV3{
		Meta:  api.NewPaginatedMeta(rs),
		Links: api.NewPaginatedLinks(reqURL, rs),
	}
	for _, r := range rs.Records {
		resp.Data = append(resp.Data, resourceV3(r))
	}

	return resp
}

func writeErrorV3(w http.ResponseWriter, err error) {
	api.WriteError(w, err)
}
