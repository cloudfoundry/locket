package presence_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPresence(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Presence Suite")
}
