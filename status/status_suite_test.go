package status_test

import (
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"testing"
)

var consulRunner *consulrunner.ClusterRunner
var consulSession *consuladapter.Session

func TestServicesBbs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status Suite")
}

var _ = BeforeSuite(func() {
	consulRunner = consulrunner.NewClusterRunner(
		5001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
		1,
		"http",
	)

	consulRunner.Start()
	consulRunner.WaitUntilReady()
})

var _ = AfterSuite(func() {
	consulRunner.Stop()
})

var _ = BeforeEach(func() {
	consulRunner.Reset()
	consulSession = consulRunner.NewSession("a-session")
})
