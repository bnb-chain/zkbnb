package types

import (
	"encoding/json"
	"fmt"
)

type Error interface {
	Error() string
	RefineError(err ...interface{}) Error
}

type BizError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func NewBusinessError(code int32, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

func (e *BizError) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (e *BizError) RefineError(err ...interface{}) Error {
	return NewBusinessError(e.Code, e.Message+fmt.Sprint(err...))
}

type SysError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func NewSystemError(code int32, message string) *SysError {
	return &SysError{
		Code:    code,
		Message: message,
	}
}

func (e *SysError) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (e *SysError) RefineError(err ...interface{}) Error {
	return NewSystemError(e.Code, e.Message+fmt.Sprint(err...))
}
