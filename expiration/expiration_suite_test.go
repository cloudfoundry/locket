package expiration_test

import (
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	sender *fake.FakeMetricSender
)

func TestExpiration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Expiration Suite")
}

var _ = BeforeSuite(func() {
	sender = fake.NewFakeMetricSender()
	metrics.Initialize(sender, nil)
})
