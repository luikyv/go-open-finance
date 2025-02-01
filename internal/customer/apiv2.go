package customer

import (
	"encoding/json"
	"net/http"

	"github.com/luikyv/go-open-finance/internal/api"
	"github.com/luikyv/go-open-finance/internal/page"
	"github.com/luikyv/go-open-finance/internal/timex"
)

type APIHandlerV2 struct {
	service Service
}

func NewAPIHandlerV2(service Service) APIHandlerV2 {
	return APIHandlerV2{
		service: service,
	}
}

func (router APIHandlerV2) GetPersonalIdentificationsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub := r.Context().Value(api.CtxKeySubject).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)
		pag, err := api.NewPagination(r)
		if err != nil {
			writeErrorV2(w, api.NewError("INVALID_PARAMETER", http.StatusUnprocessableEntity, err.Error()))
			return
		}

		identifications := router.service.personalIdentifications(r.Context(), sub, pag)
		resp := toPersonalIdentificationsResponseV2(identifications, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err)
			return
		}
	})
}

func (router APIHandlerV2) GetPersonalQualificationsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub := r.Context().Value(api.CtxKeySubject).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)

		qualifications := router.service.personalQualifications(r.Context(), sub)
		resp := toPersonalQualificationsResponseV2(qualifications, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err)
			return
		}
	})
}

func (router APIHandlerV2) GetPersonalFinancialRelationsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub := r.Context().Value(api.CtxKeySubject).(string)
		reqURL := r.Context().Value(api.CtxKeyRequestURL).(string)

		financialRelation := router.service.personalFinancialRelations(r.Context(), sub)
		resp := toPersonalFinancialRelationsResponseV2(financialRelation, reqURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeErrorV2(w, err)
			return
		}
	})
}

type personalIdentificationsResponseV2 struct {
	Data  []personalIdentificationV2 `json:"data"`
	Meta  api.Meta                   `json:"meta"`
	Links api.Links                  `json:"links"`
}

type personalIdentificationV2 struct {
	UpdateDateTime    timex.DateTime `json:"updateDateTime"`
	PersonalID        string         `json:"personalId"`
	BrandName         string         `json:"brandName"`
	CivilName         string         `json:"civilName"`
	SocialName        string         `json:"socialName,omitempty"`
	BirthDate         timex.Date     `json:"birthDate"`
	MaritalStatusCode MaritalStatus  `json:"maritalStatusCode"`
	Sex               Sex            `json:"sex"`
	CompaniesCNPJ     []string       `json:"companiesCnpj"`
	Documents         struct {
		CPF string `json:"cpfNumber"`
	} `json:"documents"`
	HasBrazilianNationality bool `json:"hasBrazilianNationality"`
	Contacts                struct {
		PostalAddresses []personalIdentificationAddressV2 `json:"postalAddresses"`
		Phones          []personalIdentificationPhoneV2   `json:"phones"`
		Emails          []personalIdentificationEmailV2   `json:"emails"`
	} `json:"contacts"`
}

type personalIdentificationAddressV2 struct {
	IsMain   bool   `json:"isMain"`
	Address  string `json:"address"`
	TownName string `json:"townName"`
	PostCode string `json:"postCode"`
	Country  string `json:"country"`
}

type personalIdentificationPhoneV2 struct {
	IsMain   bool      `json:"isMain"`
	Type     PhoneType `json:"type"`
	AreaCode string    `json:"areaCode"`
	Number   string    `json:"number"`
}

type personalIdentificationEmailV2 struct {
	IsMain bool   `json:"isMain"`
	Email  string `json:"email"`
}

func toPersonalIdentificationsResponseV2(ids page.Page[PersonalIdentification], reqURL string) personalIdentificationsResponseV2 {
	resp := personalIdentificationsResponseV2{
		Meta:  api.NewMeta(),
		Links: api.NewPaginatedLinks(reqURL, ids),
	}
	for _, id := range ids.Records {
		data := personalIdentificationV2{
			UpdateDateTime:    id.UpdateDateTime,
			PersonalID:        id.ID,
			BrandName:         id.BrandName,
			CivilName:         id.CivilName,
			SocialName:        id.SocialName,
			BirthDate:         id.BirthDate,
			MaritalStatusCode: id.MaritalStatus,
			Sex:               id.Sex,
			CompaniesCNPJ:     []string{id.CompanyCNPJ},
		}
		data.Documents.CPF = id.CPF
		for _, address := range id.Addresses {
			data.Contacts.PostalAddresses = append(data.Contacts.PostalAddresses, personalIdentificationAddressV2(address))
		}
		for _, phone := range id.Phones {
			data.Contacts.Phones = append(data.Contacts.Phones, personalIdentificationPhoneV2(phone))
		}
		for _, email := range id.Emails {
			data.Contacts.Emails = append(data.Contacts.Emails, personalIdentificationEmailV2(email))
		}

		resp.Data = append(resp.Data, data)
	}

	return resp
}

type personalQualificationsResponseV2 struct {
	Data struct {
		UpdateDateTime        timex.DateTime `json:"updateDateTime"`
		CompanyCNPJ           string         `json:"companyCnpj"`
		OccupationCode        Occupation     `json:"occupationCode,omitempty"`
		OccupationDescription string         `json:"occupationDescription,omitempty"`
	} `json:"data"`
	Meta  api.Meta  `json:"meta"`
	Links api.Links `json:"links"`
}

func toPersonalQualificationsResponseV2(qs PersonalQualifications, reqURL string) personalQualificationsResponseV2 {
	resp := personalQualificationsResponseV2{
		Data: struct {
			UpdateDateTime        timex.DateTime "json:\"updateDateTime\""
			CompanyCNPJ           string         "json:\"companyCnpj\""
			OccupationCode        Occupation     "json:\"occupationCode,omitempty\""
			OccupationDescription string         "json:\"occupationDescription,omitempty\""
		}{
			UpdateDateTime:        qs.UpdateDateTime,
			CompanyCNPJ:           qs.CompanyCNPJ,
			OccupationCode:        qs.Occupation,
			OccupationDescription: qs.OccupationDescription,
		},
		Meta: api.NewMeta(),
		Links: api.Links{
			Self: reqURL,
		},
	}

	return resp
}

type personalFinancialRelationsResponseV2 struct {
	Data struct {
		UpdateDateTime               timex.DateTime                        `json:"updateDateTime"`
		StartDate                    timex.DateTime                        `json:"startDate"`
		ProductServiceTypes          []ProductServiceType                  `json:"productsServicesType"`
		ProductServiceAdditionalInfo string                                `json:"productsServicesTypeAdditionalInfo,omitempty"`
		Procurators                  []struct{}                            `json:"procurators"`
		Accounts                     []personalFinancialRelationsAccountV2 `json:"accounts"`
	} `json:"data"`
	Meta  api.Meta  `json:"meta"`
	Links api.Links `json:"links"`
}

type personalFinancialRelationsAccountV2 struct {
	CompeCode  string         `json:"compeCode"`
	Branch     string         `json:"branchCode,omitempty"`
	Number     string         `json:"number"`
	CheckDigit string         `json:"checkDigit"`
	Type       AccountType    `json:"type"`
	SubType    AccountSubType `json:"subtype"`
}

func toPersonalFinancialRelationsResponseV2(rels PersonalFinancialRelations, reqURL string) personalFinancialRelationsResponseV2 {
	resp := personalFinancialRelationsResponseV2{
		Data: struct {
			UpdateDateTime               timex.DateTime                        "json:\"updateDateTime\""
			StartDate                    timex.DateTime                        "json:\"startDate\""
			ProductServiceTypes          []ProductServiceType                  "json:\"productsServicesType\""
			ProductServiceAdditionalInfo string                                "json:\"productsServicesTypeAdditionalInfo,omitempty\""
			Procurators                  []struct{}                            "json:\"procurators\""
			Accounts                     []personalFinancialRelationsAccountV2 "json:\"accounts\""
		}{
			UpdateDateTime:               rels.UpdateDateTime,
			StartDate:                    rels.StartDateTime,
			ProductServiceTypes:          rels.ProductServiceTypes,
			ProductServiceAdditionalInfo: rels.ProductServiceAdditionalInfo,
			Procurators:                  []struct{}{},
		},
		Meta: api.NewMeta(),
		Links: api.Links{
			Self: reqURL,
		},
	}
	for _, account := range rels.Accounts {
		resp.Data.Accounts = append(resp.Data.Accounts, personalFinancialRelationsAccountV2(account))
	}

	return resp
}

func writeErrorV2(w http.ResponseWriter, err error) {
	api.WriteError(w, err)
}
