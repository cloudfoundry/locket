package main_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/localip"
	"code.cloudfoundry.org/locket/cmd/locket/config"
	"code.cloudfoundry.org/locket/cmd/locket/testrunner"
	"code.cloudfoundry.org/locket/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/hashicorp/consul/api"
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
		locketPort    uint16
		locketRunner  ifrit.Runner

		caCertFile, certFile, keyFile string
		tlsConfig                     *tls.Config

		cfg config.LocketConfig
	)

	BeforeEach(func() {
		var err error

		locketPort, err = localip.LocalPort()
		Expect(err).NotTo(HaveOccurred())

		locketAddress = fmt.Sprintf("127.0.0.1:%d", locketPort)

		caCertFile = "fixtures/ca.crt"
		certFile = "fixtures/cert.crt"
		keyFile = "fixtures/key.key"

		tlsConfig, err = cfhttp.NewTLSConfig(certFile, keyFile, caCertFile)
		Expect(err).NotTo(HaveOccurred())

		cfg = config.LocketConfig{
			ListenAddress:            locketAddress,
			ConsulCluster:            consulRunner.ConsulCluster(),
			DatabaseDriver:           sqlRunner.DriverName(),
			DatabaseConnectionString: sqlRunner.ConnectionString(),
			CaFile:   caCertFile,
			CertFile: certFile,
			KeyFile:  keyFile,
		}
	})

	JustBeforeEach(func() {
		var err error
		locketRunner = testrunner.NewLocketRunner(locketBinPath, cfg)
		locketProcess = ginkgomon.Invoke(locketRunner)

		conn, err = grpc.Dial(locketAddress, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
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

	Context("ServiceRegistration", func() {
		It("registers itself with consul", func() {
			consulClient := consulRunner.NewClient()
			services, err := consulClient.Agent().Services()
			Expect(err).ToNot(HaveOccurred())

			Expect(services).To(HaveKeyWithValue("locket",
				&api.AgentService{
					Service: "locket",
					ID:      "locket",
					Port:    int(locketPort),
				}))
		})

		It("registers a TTL healthcheck", func() {
			consulClient := consulRunner.NewClient()
			checks, err := consulClient.Agent().Checks()
			Expect(err).ToNot(HaveOccurred())

			Expect(checks).To(HaveKeyWithValue("service:locket",
				&api.AgentCheck{
					Node:        "0",
					CheckID:     "service:locket",
					Name:        "Service 'locket' check",
					Status:      "passing",
					ServiceID:   "locket",
					ServiceName: "locket",
				}))
		})
	})

	Context("Lock", func() {
		It("locks the key with the corresponding value", func() {
			requestedResource := &models.Resource{Key: "test", Value: "test-data", Owner: "jim", Type: "lock"}
			_, err := locketClient.Lock(context.Background(), &models.LockRequest{
				Resource:     requestedResource,
				TtlInSeconds: 10,
			})
			Expect(err).NotTo(HaveOccurred())

			resp, err := locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test"})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Resource).To(BeEquivalentTo(requestedResource))

			requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "nima", Type: "lock"}
			_, err = locketClient.Lock(context.Background(), &models.LockRequest{
				Resource:     requestedResource,
				TtlInSeconds: 10,
			})
			Expect(err).To(HaveOccurred())
		})

		It("expires after a ttl", func() {
			requestedResource := &models.Resource{Key: "test", Value: "test-data", Owner: "jim", Type: "lock"}
			_, err := locketClient.Lock(context.Background(), &models.LockRequest{
				Resource:     requestedResource,
				TtlInSeconds: 1,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				_, err := locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test"})
				return err
			}, 2*time.Second).Should(HaveOccurred())
		})

		Context("when the lock server disappears unexpectedly", func() {
			It("still disappears after ~ the ttl", func() {
				requestedResource := &models.Resource{Key: "test", Value: "test-data", Owner: "jim", Type: "lock"}
				_, err := locketClient.Lock(context.Background(), &models.LockRequest{
					Resource:     requestedResource,
					TtlInSeconds: 3,
				})
				Expect(err).NotTo(HaveOccurred())

				ginkgomon.Kill(locketProcess)

				locketRunner = testrunner.NewLocketRunner(locketBinPath, cfg)
				locketProcess = ginkgomon.Invoke(locketRunner)

				// Recreate the grpc client to avoid default backoff
				conn, err = grpc.Dial(locketAddress, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
				Expect(err).NotTo(HaveOccurred())
				locketClient = models.NewLocketClient(conn)

				Eventually(func() error {
					_, err := locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test"})
					return err
				}, 6*time.Second).Should(HaveOccurred())
			})
		})
	})

	Context("Release", func() {
		var requestedResource *models.Resource

		Context("when the lock does not exist", func() {
			It("throws an error releasing the lock", func() {
				requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "jim", Type: "lock"}
				_, err := locketClient.Release(context.Background(), &models.ReleaseRequest{Resource: requestedResource})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the lock exists", func() {
			JustBeforeEach(func() {
				requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "jim", Type: "lock"}
				_, err := locketClient.Lock(context.Background(), &models.LockRequest{Resource: requestedResource, TtlInSeconds: 10})
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
					requestedResource = &models.Resource{Key: "test", Value: "test-data", Owner: "nima", Type: "lock"}
					_, err := locketClient.Release(context.Background(), &models.ReleaseRequest{Resource: requestedResource})
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Context("FetchAll", func() {
		var (
			resource1, resource2, resource3 *models.Resource
		)

		JustBeforeEach(func() {
			resource1 = &models.Resource{Key: "test", Value: "test-data", Owner: "jim", Type: "lock"}
			resource2 = &models.Resource{Key: "iamakey", Value: "test-data", Owner: "jim", Type: "presence"}
			resource3 = &models.Resource{Key: "newkey", Value: "test-data", Owner: "jim", Type: ""}

			_, err := locketClient.Lock(context.Background(), &models.LockRequest{
				Resource:     resource1,
				TtlInSeconds: 10,
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = locketClient.Lock(context.Background(), &models.LockRequest{
				Resource:     resource2,
				TtlInSeconds: 10,
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = locketClient.Lock(context.Background(), &models.LockRequest{
				Resource:     resource3,
				TtlInSeconds: 10,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("fetches all the locks", func() {
			response, err := locketClient.FetchAll(context.Background(), &models.FetchAllRequest{})
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Resources).To(ConsistOf(resource1, resource2, resource3))

			response, err = locketClient.FetchAll(context.Background(), &models.FetchAllRequest{Type: "presence"})
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Resources).To(ConsistOf(resource2))
		})
	})
})
