package models

import (
	"fmt"

	"github.com/luikyv/go-opf/gopf/constants"
)

type OPFError interface {
	Code() constants.ErrorCode
	Response() ResponseError
	error
}

type opfError struct {
	code        constants.ErrorCode
	description string
}

func (err opfError) Code() constants.ErrorCode {
	return err.code
}

func (err opfError) Response() ResponseError {
	return NewResponseError(err.code, err.description)
}

func (err opfError) Error() string {
	return fmt.Sprintf("%s %s", err.code, err.description)
}

func NewOPFError(code constants.ErrorCode, description string) OPFError {
	return opfError{
		code:        code,
		description: description,
	}
}
