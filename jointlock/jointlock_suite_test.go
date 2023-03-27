package jointlock_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestJointlock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jointlock Suite")
}
