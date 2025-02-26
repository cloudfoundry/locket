package grpcserver_test

import (
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/locket/grpcserver"
	"code.cloudfoundry.org/locket/models"
	"code.cloudfoundry.org/tlsconfig"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
	"golang.org/x/net/context"
)

var _ = Describe("GRPCServer", func() {
	var (
		logger        *lagertest.TestLogger
		listenAddress string
		runner        ifrit.Runner
		serverProcess ifrit.Process
		tlsConfig     *tls.Config

		certFixture, keyFixture, caCertFixture string
	)

	BeforeEach(func() {
		var err error

		certFixture = "fixtures/cert.crt"
		keyFixture = "fixtures/cert.key"
		caCertFixture = "fixtures/ca.crt"

		tlsConfig, err = tlsconfig.Build(
			tlsconfig.WithInternalServiceDefaults(),
			tlsconfig.WithIdentityFromFile(certFixture, keyFixture),
		).Server(tlsconfig.WithClientAuthenticationFromFile(caCertFixture))
		Expect(err).NotTo(HaveOccurred())

		logger = lagertest.NewTestLogger("grpc-server")
		port, err := portAllocator.ClaimPorts(1)
		Expect(err).NotTo(HaveOccurred())
		listenAddress = fmt.Sprintf("localhost:%d", port)

		runner = grpcserver.NewGRPCServer(logger, listenAddress, tlsConfig, &testHandler{})
	})

	JustBeforeEach(func() {
		serverProcess = ginkgomon.Invoke(runner)
	})

	AfterEach(func() {
		ginkgomon.Kill(serverProcess)
	})

	It("serves on the listen address", func() {
		clientTLSConfig, err := tlsconfig.Build(
			tlsconfig.WithInternalServiceDefaults(),
			tlsconfig.WithIdentityFromFile(certFixture, keyFixture),
		).Client(tlsconfig.WithAuthorityFromFile(caCertFixture))
		Expect(err).NotTo(HaveOccurred())

		conn, err := grpc.NewClient(listenAddress, grpc.WithTransportCredentials(credentials.NewTLS(clientTLSConfig)))
		Expect(err).NotTo(HaveOccurred())

		locketClient := models.NewLocketClient(conn)
		_, err = locketClient.Lock(context.Background(), &models.ProtoLockRequest{})
		Expect(err).NotTo(HaveOccurred())

		_, err = locketClient.Release(context.Background(), &models.ProtoReleaseRequest{})
		Expect(err).NotTo(HaveOccurred())

		_, err = locketClient.Fetch(context.Background(), &models.ProtoFetchRequest{})
		Expect(err).NotTo(HaveOccurred())

		_, err = locketClient.FetchAll(context.Background(), &models.ProtoFetchAllRequest{})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the server fails to listen", func() {
		var alternateRunner ifrit.Runner

		BeforeEach(func() {
			alternateRunner = grpcserver.NewGRPCServer(logger, listenAddress, tlsConfig, &testHandler{})
		})

		It("exits with an error", func() {
			var err error
			process := ifrit.Background(alternateRunner)
			Eventually(process.Wait()).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
		})
	})
})

type testHandler struct {
	models.UnimplementedLocketServer
}

func (h *testHandler) Lock(ctx context.Context, req *models.ProtoLockRequest) (*models.ProtoLockResponse, error) {
	return &models.ProtoLockResponse{}, nil
}
func (h *testHandler) Release(ctx context.Context, req *models.ProtoReleaseRequest) (*models.ProtoReleaseResponse, error) {
	return &models.ProtoReleaseResponse{}, nil
}
func (h *testHandler) Fetch(ctx context.Context, req *models.ProtoFetchRequest) (*models.ProtoFetchResponse, error) {
	return &models.ProtoFetchResponse{}, nil
}
func (h *testHandler) FetchAll(ctx context.Context, req *models.ProtoFetchAllRequest) (*models.ProtoFetchAllResponse, error) {
	return &models.ProtoFetchAllResponse{}, nil
}
