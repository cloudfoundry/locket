package main_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	locket "code.cloudfoundry.org/locket/cmd/locket"
	"code.cloudfoundry.org/locket/db/dbfakes"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
)

var _ = Describe("DBHealthCheckRunner", func() {
	var (
		fakeClock  *fakeclock.FakeClock
		fakeDB     *dbfakes.FakeLocketHealthCheckDB
		fakeLogger *lagertest.TestLogger
		runner     *locket.DBHealthCheckRunner
		process    ifrit.Process
	)
	BeforeEach(func() {
		fakeClock = fakeclock.NewFakeClock(time.Now())
		fakeLogger = lagertest.NewTestLogger("test")
		fakeDB = &dbfakes.FakeLocketHealthCheckDB{}
		fakeDB.PerformLocketHealthCheckReturns(nil)
		runner = locket.NewDBHealthCheckRunner(fakeLogger, fakeDB, fakeClock, 4, 100*time.Millisecond, 200*time.Millisecond)
	})
	JustBeforeEach(func() {
		process = ginkgomon.Invoke(runner)
		Eventually(fakeLogger).Should(gbytes.Say("reentering-run-loop"))
	})
	AfterEach(func() {
		ginkgomon.Kill(process)
		waitCh := process.Wait()
		Eventually(waitCh).Should(Receive())
	})

	Context("when using empty values for health check settings", func() {
		BeforeEach(func() {
			runner = locket.NewDBHealthCheckRunner(fakeLogger, fakeDB, fakeClock, 0, 0, 0)
		})
		It("sets default values", func() {
			Expect(runner.HealthCheckFailureThreshold).To(Equal(3))
			Expect(runner.HealthCheckTimeout).To(Equal(5 * time.Second))
			Expect(runner.HealthCheckInterval).To(Equal(10 * time.Second))
		})
		It("queries PerformLocketHealthCheck() every interval", func() {
			callCount := fakeDB.PerformLocketHealthCheckCallCount()
			fakeClock.IncrementBySeconds(11)
			Eventually(fakeDB.PerformLocketHealthCheckCallCount).Should(BeNumerically(">", callCount))
			fakeClock.IncrementBySeconds(11)
			Eventually(fakeDB.PerformLocketHealthCheckCallCount).Should(BeNumerically(">", callCount+1))
		})
	})

	It("queries PerformLocketHealthCheck() every interval", func() {
		callCount := fakeDB.PerformLocketHealthCheckCallCount()
		fakeClock.Increment(201 * time.Millisecond)
		Eventually(fakeDB.PerformLocketHealthCheckCallCount).Should(BeNumerically(">", callCount))
		Eventually(fakeLogger).Should(gbytes.Say("health-check-succeeded"))
		fakeClock.Increment(201 * time.Millisecond)
		Eventually(fakeDB.PerformLocketHealthCheckCallCount).Should(BeNumerically(">", callCount+1))
	})
	Context("when signaled", func() {
		It("exits without an error", func() {
			ginkgomon.Interrupt(process)
			waitCh := process.Wait()
			Eventually(func(g Gomega) {
				err := <-waitCh
				g.Expect(err).ToNot(HaveOccurred())
			}).Should(Succeed())
			Consistently(fakeLogger).ShouldNot(gbytes.Say("database-failure-detected-restarting-locket"))
		})
	})
	Context("and the health check is hanging up to the timeout", func() {
		BeforeEach(func() {
			runner.HealthCheckTimeout = 200 * time.Millisecond
			runner.HealthCheckInterval = 100 * time.Millisecond
			fakeDB.PerformLocketHealthCheckCalls(func(ctx context.Context, logger lager.Logger, t time.Time) error {
				time.Sleep(200 * time.Second)
				return nil
			})
		})
		It("only runs one health check at a time", func() {
			waitCh := process.Wait()
			fakeClock.Increment(100 * time.Millisecond)
			time.Sleep(100 * time.Millisecond)
			fakeClock.Increment(100 * time.Millisecond)
			Consistently(waitCh, "300ms").ShouldNot(Receive())
			Expect(fakeDB.PerformLocketHealthCheckCallCount()).To(Equal(1))
		})
		Context("and it is signalled", func() {
			It("processes the signal immediately", func() {
				ginkgomon.Interrupt(process)
				waitCh := process.Wait()
				Eventually(func(g Gomega) {
					err := <-waitCh
					g.Expect(err).ToNot(HaveOccurred())
				}, "100ms").Should(Succeed())

			})
		})
	})
	Context("ExecuteTimedHealthCheckWithRetries()", func() {
		Context("when PerformLocketHealthCheck() fails", func() {
			Context("the entire time", func() {
				BeforeEach(func() {
					fakeDB.PerformLocketHealthCheckReturns(fmt.Errorf("meow"))
				})
				It("returns an error", func() {
					fakeClock.IncrementBySeconds(2)
					waitCh := process.Wait()
					Eventually(waitCh).Should(Receive(MatchError("meow\nmeow\nmeow\nmeow")))
					Eventually(fakeLogger).Should(gbytes.Say("database-failure-detected-restarting-locket"))
					Expect(fakeDB.PerformLocketHealthCheckCallCount()).To(Equal(4))

				})
			})
			Context("only twice and then succeeds", func() {
				BeforeEach(func() {
					fakeDB.PerformLocketHealthCheckReturnsOnCall(0, fmt.Errorf("meow"))
					fakeDB.PerformLocketHealthCheckReturnsOnCall(1, fmt.Errorf("meow"))
					fakeDB.PerformLocketHealthCheckReturnsOnCall(2, nil)
				})
				It("returns a success", func() {
					fakeClock.IncrementBySeconds(15)
					waitCh := process.Wait()
					Consistently(waitCh).ShouldNot(Receive())
					Eventually(fakeLogger).Should(gbytes.Say("health-check-succeeded"))
				})
			})
		})
	})
	Describe("ExecuteTimedHealthCheck()", func() {
		Context("when PerformLocketHealthCheck returns an error", func() {
			BeforeEach(func() {
				fakeDB.PerformLocketHealthCheckReturns(fmt.Errorf("meow"))
			})
			It("fails", func() {
				err := runner.ExecuteTimedHealthCheck()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("meow"))
			})
		})
		Context("when PerformLocketHealthCheck succeeds", func() {
			It("succeeds", func() {
				err := runner.ExecuteTimedHealthCheck()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("when PerformLocketHealthCheck takes > the timeout", func() {
			BeforeEach(func() {
				fakeDB.PerformLocketHealthCheckCalls(func(ctx context.Context, logger lager.Logger, t time.Time) error {
					fakeClock.Increment(200 * time.Millisecond)
					time.Sleep(200 * time.Millisecond)
					return nil
				})
			})

			It("returns an error", func() {
				err := runner.ExecuteTimedHealthCheck()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("timed out after 100ms while executing DB health check"))

			})
			Context("when using the default timeout", func() {
				BeforeEach(func() {
					runner = locket.NewDBHealthCheckRunner(fakeLogger, fakeDB, fakeClock, 0, 0, 0)
					fakeDB.PerformLocketHealthCheckCalls(func(ctx context.Context, logger lager.Logger, t time.Time) error {
						fakeClock.Increment(6 * time.Second)
						time.Sleep(200 * time.Millisecond)
						return nil
					})
				})
				It("returns an error", func() {
					err := runner.ExecuteTimedHealthCheck()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("timed out after 5s while executing DB health check"))
				})
			})
		})
	})
})
