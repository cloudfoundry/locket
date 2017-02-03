package lock_test

import (
	"errors"
	"time"

	"google.golang.org/grpc"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/locket/lock"
	"code.cloudfoundry.org/locket/models"
	"code.cloudfoundry.org/locket/models/modelsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"golang.org/x/net/context"
)

var _ = Describe("Lock", func() {
	var (
		logger *lagertest.TestLogger

		fakeLocker *modelsfakes.FakeLocketClient
		fakeClock  *fakeclock.FakeClock

		expectedLock      *models.Resource
		lockRetryInterval time.Duration

		lockRunner  ifrit.Runner
		lockProcess ifrit.Process
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("lock")

		fakeLocker = &modelsfakes.FakeLocketClient{}
		fakeClock = fakeclock.NewFakeClock(time.Now())

		lockRetryInterval = locket.RetryInterval
		expectedLock = &models.Resource{Key: "test", Owner: "jim", Value: "is pretty sweet."}

		lockRunner = lock.NewLockRunner(
			logger,
			fakeLocker,
			expectedLock,
			fakeClock,
			lockRetryInterval,
		)
	})

	JustBeforeEach(func() {
		lockProcess = ifrit.Background(lockRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(lockProcess)
	})

	It("locks the key", func() {
		Eventually(lockProcess.Ready()).Should(BeClosed())
		Eventually(fakeLocker.LockCallCount).Should(Equal(1))
		_, lockReq, _ := fakeLocker.LockArgsForCall(0)
		Expect(lockReq.Resource).To(Equal(expectedLock))
	})

	Context("when the lock cannot be acquired", func() {
		BeforeEach(func() {
			fakeLocker.LockReturns(nil, errors.New("no-lock-for-you"))
		})

		It("retries locking after the lock retry interval", func() {
			Eventually(fakeLocker.LockCallCount).Should(Equal(1))
			_, lockReq, _ := fakeLocker.LockArgsForCall(0)
			Expect(lockReq.Resource).To(Equal(expectedLock))

			fakeClock.WaitForWatcherAndIncrement(lockRetryInterval)

			Eventually(fakeLocker.LockCallCount).Should(Equal(2))
			_, lockReq, _ = fakeLocker.LockArgsForCall(1)
			Expect(lockReq.Resource).To(Equal(expectedLock))

			Consistently(lockProcess.Ready()).ShouldNot(BeClosed())
		})

		Context("and the lock becomes available", func() {
			var done chan struct{}

			BeforeEach(func() {
				done = make(chan struct{})

				fakeLocker.LockStub = func(ctx context.Context, res *models.LockRequest, opts ...grpc.CallOption) (*models.LockResponse, error) {
					select {
					case <-done:
						return nil, nil
					default:
						return nil, errors.New("no-lock-for-you")
					}
				}
			})

			It("stops retrying to grab the lock", func() {
				Eventually(fakeLocker.LockCallCount).Should(Equal(1))
				_, lockReq, _ := fakeLocker.LockArgsForCall(0)
				Expect(lockReq.Resource).To(Equal(expectedLock))
				Consistently(lockProcess.Ready()).ShouldNot(BeClosed())

				close(done)
				fakeClock.WaitForWatcherAndIncrement(lockRetryInterval)

				Eventually(lockProcess.Ready()).Should(BeClosed())
				Eventually(fakeLocker.LockCallCount).Should(Equal(2))
				_, lockReq, _ = fakeLocker.LockArgsForCall(1)
				Expect(lockReq.Resource).To(Equal(expectedLock))

				Consistently(fakeClock.WatcherCount()).Should(Equal(0))
				fakeClock.Increment(lockRetryInterval)
				Consistently(fakeLocker.LockCallCount).Should(Equal(2))
			})
		})
	})

	Context("when the lock can be acquired", func() {
		It("grabs the lock and then stops trying to grab it", func() {
			Eventually(lockProcess.Ready()).Should(BeClosed())
			Eventually(fakeLocker.LockCallCount).Should(Equal(1))
			_, lockReq, _ := fakeLocker.LockArgsForCall(0)
			Expect(lockReq.Resource).To(Equal(expectedLock))

			Consistently(fakeClock.WatcherCount()).Should(Equal(0))
			fakeClock.Increment(lockRetryInterval)
			Consistently(fakeLocker.LockCallCount).Should(Equal(1))
		})
	})

	Context("when the lock process receives a signal", func() {
		It("releases the lock", func() {
			ginkgomon.Interrupt(lockProcess)
			Eventually(fakeLocker.ReleaseCallCount).Should(Equal(1))
			_, releaseReq, _ := fakeLocker.ReleaseArgsForCall(0)
			Expect(releaseReq.Resource).To(Equal(expectedLock))
		})
	})
})
