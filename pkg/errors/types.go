package errors

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Reason   string `json:"reason,omitempty"`
	Message  string `json:"message,omitempty"`
	RawError string `json:"rawError,omitempty"`
	Raw      error  `json:"-"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

type ErrBuilder = Error

func New() *ErrBuilder {
	return &ErrBuilder{}
}
func (e *ErrBuilder) SetReason(reason string) *ErrBuilder {
	e.Reason = reason
	return e
}
func (e *ErrBuilder) SetMessage(msg string) *ErrBuilder {
	e.Message = msg
	return e
}
func (e *ErrBuilder) SetFormatMsg(format string, v ...interface{}) *ErrBuilder {
	e.Message = fmt.Sprintf(format, v...)
	return e
}
func (e *ErrBuilder) SetRaw(err error) *ErrBuilder {
	e.Raw = err
	if err != nil {
		e.RawError = err.Error()
	} else {
		e.RawError = ""
	}
	return e
}
func (e *ErrBuilder) Get() *Error {
	return (*Error)(e)
}
