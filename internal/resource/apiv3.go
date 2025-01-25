package resource

import (
	"encoding/json"
	"net/http"

	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/page"
)

type APIHandlerV3 struct {
	service Service
}

func NewAPIHandlerV3(service Service) APIHandlerV3 {
	return APIHandlerV3{
		service: service,
	}
}

func (router APIHandlerV3) GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		consentID := r.Context().Value(api.CtxKeyConsentID).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		pag := api.NewPagination(r)

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
		Meta:  api.NewMeta(),
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
