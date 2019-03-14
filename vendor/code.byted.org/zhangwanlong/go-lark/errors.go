package lark

import (
	"errors"
	"reflect"
)

// Lark API Errors
const (
	ErrSuccess        = 0
	ErrInvalidParams  = 2
	ErrNotExist       = 3
	ErrForbidden      = 4
	ErrLoginRequired  = 5
	ErrExceedLimit    = 6
	ErrDeleted        = 7
	ErrNotInRoom      = 8
	ErrUnchanged      = 9
	ErrWrongConn      = 10
	ErrIPForbidden    = 11
	ErrFrequencyLimit = 12
	ErrUnauth         = 13
	ErrInternalError  = 14
	ErrDenoised       = 15
)

func validResponse(resp interface{}) bool {
	if resp == nil {
		return false
	}
	val := reflect.ValueOf(resp)
	if val.IsNil() {
		return false
	}
	ok := val.Elem().FieldByName("Ok")
	code := val.Elem().FieldByName("Code")
	err := val.Elem().FieldByName("Error")
	if !ok.IsValid() || !code.IsValid() || !err.IsValid() {
		return false
	}
	if ok.Kind() != reflect.Bool || code.Kind() != reflect.Int || err.Kind() != reflect.String {
		return false
	}
	return true
}

// IsError check whether there is a response error
func IsError(resp interface{}) bool {
	if !validResponse(resp) {
		return true
	}
	ok := reflect.ValueOf(resp).Elem().FieldByName("Ok")
	if ok.Bool() {
		return false
	}
	return true
}

// GetError returns error
// See https://docs.bytedance.net/doc/1NVmJnaDi9l2BRfC6FZdSd for further explanation.
func GetError(resp interface{}) error {
	if !validResponse(resp) {
		return errors.New("Invalid response")
	}
	ok := reflect.ValueOf(resp).Elem().FieldByName("Ok")
	if ok.Bool() {
		return nil
	}

	code := reflect.ValueOf(resp).Elem().FieldByName("Code").Int()
	errmsg := reflect.ValueOf(resp).Elem().FieldByName("Error").String()
	var err error
	if code == ErrSuccess {
		err = nil
	} else if code >= 2 && code <= 15 {
		err = errors.New(errmsg)
	} else {
		err = errors.New("Unknown error")
	}
	return err
}
