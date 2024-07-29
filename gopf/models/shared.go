package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luikyv/go-opf/gopf/constants"
)

type DateTime time.Time

func (d DateTime) MarshalJSON() ([]byte, error) {
	jsonStr := "\"" + time.Time(d).Format(constants.RFC3339) + "\""
	return []byte(jsonStr), nil
}

func (d *DateTime) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return errors.New("not a json string")
	}

	// Strip the double quotes from the JSON string.
	b = b[1 : len(b)-1]

	// Parse the result using date time format.
	t, err := time.Parse(constants.RFC3339, string(b))
	if err != nil {
		return fmt.Errorf("failed to parse time: %w", err)
	}

	*d = DateTime(t)
	return nil
}

func DateTimeNow() DateTime {
	return DateTime(time.Now())
}

func (d DateTime) Unix() int64 {
	return time.Time(d).Unix()
}

func DateTimeUnix(timestamp int64) DateTime {
	return DateTime(time.Unix(timestamp, 0))
}

type Document struct {
	Identification string `json:"identification"`
	Rel            string `json:"rel"`
}

type LoggedUser struct {
	Document Document `json:"document"`
}

type BusinessEntity struct {
	Document Document `json:"document"`
}

type Links struct {
	Self string `json:"self"`
}

type Meta struct {
	RequestDateTime DateTime `json:"requestDateTime"`
}

type Response struct {
	Data  any   `json:"data"`
	Meta  Meta  `json:"meta"`
	Links Links `json:"links"`
}

type Error struct {
	Code   constants.ErrorCode `json:"code"`
	Title  string              `json:"title"`
	Detail string              `json:"detail"`
}

type ResponseError struct {
	Errors []Error `json:"errors"`
	Meta   Meta
}

func NewResponseError(code constants.ErrorCode, description string) ResponseError {
	return ResponseError{
		Errors: []Error{{Code: code, Title: description, Detail: description}},
		Meta: Meta{
			RequestDateTime: DateTimeNow(),
		},
	}
}

type CallerInfo struct {
	ClientID string
}

func NewCallerInfo(ctx *gin.Context) CallerInfo {
	return CallerInfo{
		ClientID: ctx.GetString(constants.CtxKeyClientID),
	}
}
