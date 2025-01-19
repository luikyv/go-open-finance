package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/luikyv/go-open-finance/internal/timex"
)

type Error struct {
	Code        string
	StatusCode  int
	Description string
}

func (err Error) Error() string {
	return fmt.Sprintf("%s %s", err.Code, err.Description)
}

func NewError(code string, status int, description string) Error {
	err := Error{
		Code:        code,
		StatusCode:  status,
		Description: description,
	}

	return err
}

func WriteError(w http.ResponseWriter, err error) {
	var apiErr Error
	if !errors.As(err, &apiErr) {
		WriteError(w, Error{"INTERNAL_ERROR", http.StatusInternalServerError, "internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	_ = json.NewEncoder(w).Encode(response{
		Errors: []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		}{
			{
				Code:   apiErr.Code,
				Title:  apiErr.Code,
				Detail: apiErr.Description,
			},
		},
	})
}

type response struct {
	Errors []struct {
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
	Meta struct {
		RequestDateTime timex.DateTime `json:"requestDateTime"`
	} `json:"meta"`
}
