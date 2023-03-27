package lock_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lock Suite")
}
