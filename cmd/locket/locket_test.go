package main_test

import (
	"context"
	"fmt"
	"net"

	"code.cloudfoundry.org/localip"
	"code.cloudfoundry.org/locket/cmd/locket/config"
	"code.cloudfoundry.org/locket/cmd/locket/testrunner"
	"code.cloudfoundry.org/locket/models"
	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Locket", func() {
	var (
		conn *grpc.ClientConn

		locketAddress string
		locketClient  models.LocketClient
		locketProcess ifrit.Process
		cfg           config.LocketConfig
	)

	BeforeEach(func() {
		locketPort, err := localip.LocalPort()
		Expect(err).NotTo(HaveOccurred())

		locketAddress = fmt.Sprintf("127.0.0.1:%d", locketPort)

		cfg = config.LocketConfig{
			ListenAddress:            locketAddress,
			DatabaseDriver:           sqlRunner.DriverName(),
			DatabaseConnectionString: sqlRunner.ConnectionString(),
		}
	})

	JustBeforeEach(func() {
		var err error
		locketRunner := testrunner.NewLocketRunner(locketBinPath, cfg)
		locketProcess = ginkgomon.Invoke(locketRunner)

		conn, err = grpc.Dial(locketAddress, grpc.WithInsecure())
		Expect(err).NotTo(HaveOccurred())

		locketClient = models.NewLocketClient(conn)
	})

	AfterEach(func() {
		Expect(conn.Close()).To(Succeed())
		ginkgomon.Kill(locketProcess)
		sqlRunner.ResetTables(TruncateTableList)
	})

	Context("debug address", func() {
		var debugAddress string

		BeforeEach(func() {
			port, err := localip.LocalPort()
			Expect(err).NotTo(HaveOccurred())

			debugAddress = fmt.Sprintf("127.0.0.1:%d", port)
			cfg.DebugAddress = debugAddress
		})

		It("listens on the debug address specified", func() {
			_, err := net.Dial("tcp", debugAddress)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when locking", func() {
		Context("Lock", func() {
			It("locks the key with the corresponding value", func() {
				requestedResource := &models.Resource{Key: "test", Value: "test-data", Owner: "jim"}
				_, err := locketClient.Lock(context.Background(), &models.LockRequest{Resource: requestedResource})
				Expect(err).NotTo(HaveOccurred())

				resp, err := locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test"})
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.Resource).To(BeEquivalentTo(requestedResource))

				requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "nima"}
				_, err = locketClient.Lock(context.Background(), &models.LockRequest{Resource: requestedResource})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Release", func() {
			var requestedResource *models.Resource

			Context("when the lock does not exist", func() {
				It("throws an error releasing the lock", func() {
					requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "jim"}
					_, err := locketClient.Release(context.Background(), &models.ReleaseRequest{Resource: requestedResource})
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when the lock exists", func() {
				JustBeforeEach(func() {
					requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "jim"}
					_, err := locketClient.Lock(context.Background(), &models.LockRequest{Resource: requestedResource})
					Expect(err).NotTo(HaveOccurred())

					resp, err := locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test"})
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.Resource).To(BeEquivalentTo(requestedResource))
				})

				It("releases the lock", func() {
					_, err := locketClient.Release(context.Background(), &models.ReleaseRequest{Resource: requestedResource})
					Expect(err).NotTo(HaveOccurred())

					_, err = locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test"})
					Expect(err).To(HaveOccurred())
				})

				Context("when another process is the lock owner", func() {
					It("throws an error", func() {
						requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "nima"}
						_, err := locketClient.Release(context.Background(), &models.ReleaseRequest{Resource: requestedResource})
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
