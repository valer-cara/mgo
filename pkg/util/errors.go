package util

import (
	"errors"
)

func AggregateErrors(errs []error) error {
	var strErrors string
	for _, err := range errs {
		strErrors = strErrors + err.Error() + "\n"
	}
	return errors.New(strErrors)
}
