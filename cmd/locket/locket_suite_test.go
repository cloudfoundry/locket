package main_test

import (
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/grpclog"

	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
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

	TruncateTableList = []string{"locks"}
	portAllocator     portauthority.PortAllocator
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

		grpclog.SetLogger(log.New(io.Discard, "", 0))

		locketBinPath = string(locketBinPathData)
		SetDefaultEventuallyTimeout(15 * time.Second)

		dbName := fmt.Sprintf("diego_%d", GinkgoParallelProcess())
		sqlRunner = test_helpers.NewSQLRunner(dbName)
		sqlProcess = ginkgomon.Invoke(sqlRunner)
	},
)

var _ = SynchronizedAfterSuite(func() {
	ginkgomon.Kill(sqlProcess)
}, func() {
	gexec.CleanupBuildArtifacts()
})
