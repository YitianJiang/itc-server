package tos

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrChecksum        = errors.New("missmatch md5")
	ErrContentTooSmall = errors.New("content too small")
)

type ErrRes struct {
	Success string `json:"success"`
	Err     struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	RemoteAddr string `json:"-"`
	RequestID  string `json:"-"`
}

func (e ErrRes) Error() string {
	return fmt.Sprintf("code=%d message=%s remoteAddr=%s reqID=%s", e.Err.Code, e.Err.Message, e.RemoteAddr, e.RequestID)
}

func DecodeErr(res *http.Response) error {
	decoder := json.NewDecoder(res.Body)
	res.Body.Close()
	errRes := new(ErrRes)
	if err := decoder.Decode(errRes); err != nil {
		errRes.Err.Code = res.StatusCode
		errRes.Err.Message = http.StatusText(res.StatusCode)
	}
	errRes.RequestID = res.Header.Get("X-Tos-Request-Id")
	errRes.RemoteAddr = res.Request.Host
	return errRes
}

func IsNotFound(err error) bool {
	er, ok := err.(*ErrRes)
	if ok && er.Err.Code == 404 {
		return true
	}
	return false
}
