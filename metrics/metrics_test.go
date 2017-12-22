package metrics_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers/helpersfakes"
	"code.cloudfoundry.org/clock/fakeclock"
	mfakes "code.cloudfoundry.org/diego-logging-client/testhelpers"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket/metrics"
	"code.cloudfoundry.org/locket/metrics/metricsfakes"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Metrics", func() {
	type FakeGauge struct {
		Name  string
		Value int
	}

	var (
		runner           ifrit.Runner
		process          ifrit.Process
		fakeMetronClient *mfakes.FakeIngressClient
		logger           *lagertest.TestLogger
		fakeClock        *fakeclock.FakeClock
		metricsInterval  time.Duration
		lockDBMetrics    *metricsfakes.FakeLockDBMetrics
		queryMonitor     *helpersfakes.FakeQueryMonitor
		metricsChan      chan FakeGauge
	)

	metricsChan = make(chan FakeGauge, 100)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("metrics")
		fakeMetronClient = new(mfakes.FakeIngressClient)
		fakeClock = fakeclock.NewFakeClock(time.Now())
		metricsInterval = 10 * time.Second

		lockDBMetrics = &metricsfakes.FakeLockDBMetrics{}
		queryMonitor = &helpersfakes.FakeQueryMonitor{}

		lockDBMetrics.CountStub = func(l lager.Logger, lockType string) (int, error) {
			switch {
			case lockType == models.LockType:
				return 3, nil
			case lockType == models.PresenceType:
				return 2, nil
			default:
				return 0, errors.New("unknown type")
			}
		}

		fakeMetronClient.SendMetricStub = func(name string, value int, opts ...loggregator.EmitGaugeOption) error {
			defer GinkgoRecover()

			Eventually(metricsChan).Should(BeSent(FakeGauge{name, value}))
			return nil
		}
		fakeMetronClient.SendDurationStub = func(name string, value time.Duration, opts ...loggregator.EmitGaugeOption) error {
			defer GinkgoRecover()

			Eventually(metricsChan).Should(BeSent(FakeGauge{name, int(value)}))
			return nil
		}
	})

	JustBeforeEach(func() {
		runner = metrics.NewMetricsNotifier(
			logger,
			fakeClock,
			fakeMetronClient,
			metricsInterval,
			lockDBMetrics,
			queryMonitor,
		)
		process = ifrit.Background(runner)
		Eventually(process.Ready()).Should(BeClosed())
	})

	AfterEach(func() {
		ginkgomon.Interrupt(process)
	})

	Context("when there are no errors retrieving counts from database", func() {
		BeforeEach(func() {
			lockDBMetrics.OpenConnectionsReturns(100)
			queryMonitor.QueriesInFlightReturns(5)
			queryMonitor.QueriesTotalReturns(105)
			queryMonitor.QueriesSucceededReturns(90)
			queryMonitor.QueriesFailedReturns(10)
			queryMonitor.ReadAndResetQueryDurationMaxReturns(time.Second)
		})

		JustBeforeEach(func() {
			fakeClock.Increment(metricsInterval)
		})

		It("emits a metric for the number of active locks", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"ActiveLocks", 3})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"ActiveLocks", 3})))
		})

		It("emits a metric for the number of active presences", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"ActivePresences", 2})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"ActivePresences", 2})))
		})

		It("emits a metric for the number of open database connections", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBOpenConnections", 100})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBOpenConnections", 100})))
		})

		It("emits a metric for the number of total queries against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesTotal", 105})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesTotal", 105})))
		})

		It("emits a metric for the number of queries succeeded against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesSucceeded", 90})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesSucceeded", 90})))
		})

		It("emits a metric for the number of queries failed against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesFailed", 10})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesFailed", 10})))
		})

		It("emits a metric for the number of queries in flight against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesInFlight", 5})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesInFlight", 5})))
		})

		It("emits a metric for the max duration of queries", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueryDurationMax", int(time.Second)})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueryDurationMax", int(time.Second)})))
		})
	})

	Context("when there are errors retrieving counts from database", func() {
		BeforeEach(func() {
			lockDBMetrics.CountReturns(1, errors.New("DB error"))
			lockDBMetrics.OpenConnectionsReturns(100)
			queryMonitor.QueriesInFlightReturns(5)
			queryMonitor.QueriesTotalReturns(105)
			queryMonitor.QueriesSucceededReturns(90)
			queryMonitor.QueriesFailedReturns(10)
			queryMonitor.ReadAndResetQueryDurationMaxReturns(time.Second)
		})

		JustBeforeEach(func() {
			fakeClock.Increment(metricsInterval)
		})

		It("emits a metric for the number of open database connections", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBOpenConnections", 100})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBOpenConnections", 100})))
		})

		It("emits a metric for the number of total queries against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesTotal", 105})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesTotal", 105})))
		})

		It("emits a metric for the number of queries succeeded against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesSucceeded", 90})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesSucceeded", 90})))
		})

		It("emits a metric for the number of queries failed against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesFailed", 10})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesFailed", 10})))
		})

		It("emits a metric for the number of queries in flight against the database", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesInFlight", 5})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueriesInFlight", 5})))
		})

		It("emits a metric for the max duration of queries", func() {
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueryDurationMax", int(time.Second)})))
			fakeClock.Increment(metricsInterval)
			Eventually(metricsChan).Should(Receive(Equal(FakeGauge{"DBQueryDurationMax", int(time.Second)})))
		})

		It("does not emit a metric for the number of active locks", func() {
			Eventually(func() int { return len(metricsChan) }).Should(BeNumerically(">=", 5))
			for i := 0; i < fakeMetronClient.SendMetricCallCount(); i++ {
				args, _, _ := fakeMetronClient.SendMetricArgsForCall(i)
				Expect(args).NotTo(Equal("ActiveLocks"))
			}
		})

		It("does not emit a metric for the number of active presences", func() {
			Eventually(func() int { return len(metricsChan) }).Should(BeNumerically(">=", 5))
			for i := 0; i < fakeMetronClient.SendMetricCallCount(); i++ {
				args, _, _ := fakeMetronClient.SendMetricArgsForCall(i)
				Expect(args).NotTo(Equal("ActiveLocks"))
			}
		})
	})
})
