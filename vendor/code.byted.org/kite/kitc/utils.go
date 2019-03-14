package kitc

import (
	"errors"
	"strings"
)

func joinErrs(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	s := make([]string, len(errs))
	for i, e := range errs {
		s[i] = e.Error()
	}
	return errors.New(strings.Join(s, ","))
}

// see comments in kitc.client.go
type _lbKey struct{}

var lbKey _lbKey
