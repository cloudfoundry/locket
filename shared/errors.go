package shared

import (
	"errors"

	"github.com/cloudfoundry-incubator/consuladapter"
)

var ErrServiceUnavailable = errors.New("service unavailable")
var ErrStoreResourceNotFound = errors.New("the requested resource could not be found in the store")

func ConvertConsulError(originalErr error) error {
	switch originalErr.(type) {
	case consuladapter.KeyNotFoundError:
		return ErrStoreResourceNotFound
	case consuladapter.PrefixNotFoundError:
		return ErrStoreResourceNotFound
	default:
		return originalErr
	}
}
