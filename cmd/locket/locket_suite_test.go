package main_test

import (
	"fmt"
	"io"
	"path"
	"time"

	"google.golang.org/grpc/grpclog"

	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	"code.cloudfoundry.org/inigo/helpers/portauthority"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	"testing"
)

var (
	locketBinPath string

	sqlProcess ifrit.Process
	sqlRunner  sqlrunner.SQLRunner

	testMetricsChan   chan *loggregator_v2.Envelope
	signalMetricsChan chan struct{}
	testIngressServer *testhelpers.TestIngressServer

	TruncateTableList                                       = []string{"locks"}
	portAllocator                                           portauthority.PortAllocator
	metronCAFile, metronServerCertFile, metronServerKeyFile string
)

func TestLocket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Locket Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		locketBinPathData, err := gexec.Build("code.cloudfoundry.org/locket/cmd/locket", "-race")
		Expect(err).NotTo(HaveOccurred())
		return []byte(locketBinPathData)
	},
	func(locketBinPathData []byte) {
		node := GinkgoParallelProcess()
		startPort := 1050 * node
		portRange := 1000
		endPort := startPort + portRange

		var err error
		portAllocator, err = portauthority.New(startPort, endPort)
		Expect(err).NotTo(HaveOccurred())

		grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, io.Discard, io.Discard))

		locketBinPath = string(locketBinPathData)
		SetDefaultEventuallyTimeout(15 * time.Second)

		dbName := fmt.Sprintf("diego_%d", GinkgoParallelProcess())
		sqlRunner = test_helpers.NewSQLRunner(dbName)
		sqlProcess = ginkgomon.Invoke(sqlRunner)
	},
)

var _ = BeforeEach(func() {

	fixturesPath := "fixtures"

	var err error
	metronCAFile = path.Join(fixturesPath, "metron", "CA.crt")
	metronServerCertFile = path.Join(fixturesPath, "metron", "metron.crt")
	metronServerKeyFile = path.Join(fixturesPath, "metron", "metron.key")
	testIngressServer, err = testhelpers.NewTestIngressServer(metronServerCertFile, metronServerKeyFile, metronCAFile)
	Expect(err).NotTo(HaveOccurred())
	receiversChan := testIngressServer.Receivers()
	testIngressServer.Start()

	testMetricsChan, signalMetricsChan = testhelpers.TestMetricChan(receiversChan)

})

var _ = AfterEach(func() {
	testIngressServer.Stop()
	close(signalMetricsChan)
})

var _ = SynchronizedAfterSuite(func() {
	ginkgomon.Kill(sqlProcess)
}, func() {
	gexec.CleanupBuildArtifacts()
})
