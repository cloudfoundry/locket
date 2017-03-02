package grpcserver_test

import (
	"fmt"

	"google.golang.org/grpc"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket/grpcserver"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"golang.org/x/net/context"
)

var _ = Describe("GRPCServer", func() {
	var (
		logger        *lagertest.TestLogger
		listenAddress string
		runner        ifrit.Runner
		serverProcess ifrit.Process
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("grpc-server")
		listenAddress = fmt.Sprintf("localhost:%d", 10000+GinkgoParallelNode())
		runner = grpcserver.NewGRPCServer(logger, listenAddress, &testHandler{})
	})

	JustBeforeEach(func() {
		serverProcess = ginkgomon.Invoke(runner)
	})

	AfterEach(func() {
		ginkgomon.Kill(serverProcess)
	})

	It("serves on the listen address", func() {
		conn, err := grpc.Dial(listenAddress, grpc.WithInsecure())
		Expect(err).NotTo(HaveOccurred())
		locketClient := models.NewLocketClient(conn)

		_, err = locketClient.Lock(context.Background(), &models.LockRequest{})
		Expect(err).NotTo(HaveOccurred())

		_, err = locketClient.Release(context.Background(), &models.ReleaseRequest{})
		Expect(err).NotTo(HaveOccurred())

		_, err = locketClient.Fetch(context.Background(), &models.FetchRequest{})
		Expect(err).NotTo(HaveOccurred())

		_, err = locketClient.FetchAll(context.Background(), &models.FetchAllRequest{})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the server fails to listen", func() {
		var alternateRunner ifrit.Runner

		BeforeEach(func() {
			alternateRunner = grpcserver.NewGRPCServer(logger, listenAddress, &testHandler{})
		})

		It("exits with an error", func() {
			var err error
			process := ifrit.Background(alternateRunner)
			Eventually(process.Wait()).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
		})
	})
})

type testHandler struct{}

func (h *testHandler) Lock(ctx context.Context, req *models.LockRequest) (*models.LockResponse, error) {
	return &models.LockResponse{}, nil
}
func (h *testHandler) Release(ctx context.Context, req *models.ReleaseRequest) (*models.ReleaseResponse, error) {
	return &models.ReleaseResponse{}, nil
}
func (h *testHandler) Fetch(ctx context.Context, req *models.FetchRequest) (*models.FetchResponse, error) {
	return &models.FetchResponse{}, nil
}
func (h *testHandler) FetchAll(ctx context.Context, req *models.FetchAllRequest) (*models.FetchAllResponse, error) {
	return &models.FetchAllResponse{}, nil
}
