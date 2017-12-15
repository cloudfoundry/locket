package metrics

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/clock"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/lager"
)

type RequestType int

const (
	requestsStartedMetric     = "RequestsStarted"
	requestsSucceededMetric   = "RequestsSucceeded"
	requestsFailedMetric      = "RequestsFailed"
	requestsInFlightMetric    = "RequestsInFlight"
	requestLatencyMaxDuration = "RequestLatencyMax"
)

type requestMetric struct {
	RequestsStarted   uint64
	RequestsSucceeded uint64
	RequestsFailed    uint64
	RequestsInFlight  uint64
	RequestLatencyMax int64
}

type metrics struct {
	lock, release, fetch, fetchAll requestMetric
}

//go:generate counterfeiter . RequestMetrics
type RequestMetrics interface {
	IncrementRequestsStartedCounter(requestType string, delta int)
	IncrementRequestsSucceededCounter(requestType string, delta int)
	IncrementRequestsFailedCounter(requestType string, delta int)
	IncrementRequestsInFlightCounter(requestType string, delta int)
	DecrementRequestsInFlightCounter(requestType string, delta int)
	UpdateLatency(requestType string, dur time.Duration)
}

type RequestMetricsNotifier struct {
	logger          lager.Logger
	clock           clock.Clock
	metricsInterval time.Duration
	metronClient    loggingclient.IngressClient
	metrics         metrics
}

func NewRequestMetricsNotifier(logger lager.Logger, ticker clock.Clock, metronClient loggingclient.IngressClient, metricsInterval time.Duration) *RequestMetricsNotifier {
	return &RequestMetricsNotifier{
		logger:          logger,
		clock:           ticker,
		metricsInterval: metricsInterval,
		metronClient:    metronClient,
	}
}

func (notifier *RequestMetricsNotifier) requestMetricForType(requestType string) *requestMetric {
	switch requestType {
	case "Lock":
		return &notifier.metrics.lock
	case "Release":
		return &notifier.metrics.release
	case "Fetch":
		return &notifier.metrics.fetch
	case "FetchAll":
		return &notifier.metrics.fetchAll
	default:
		panic(fmt.Sprintf("unknown request type %s", requestType))
	}
}

func (notifier *RequestMetricsNotifier) IncrementRequestsStartedCounter(requestType string, delta int) {
	atomic.AddUint64(&notifier.requestMetricForType(requestType).RequestsStarted, uint64(delta))
}

func (notifier *RequestMetricsNotifier) IncrementRequestsSucceededCounter(requestType string, delta int) {
	atomic.AddUint64(&notifier.requestMetricForType(requestType).RequestsSucceeded, uint64(delta))
}

func (notifier *RequestMetricsNotifier) IncrementRequestsFailedCounter(requestType string, delta int) {
	atomic.AddUint64(&notifier.requestMetricForType(requestType).RequestsFailed, uint64(delta))
}

func (notifier *RequestMetricsNotifier) IncrementRequestsInFlightCounter(requestType string, delta int) {
	atomic.AddUint64(&notifier.requestMetricForType(requestType).RequestsInFlight, uint64(delta))
}

func (notifier *RequestMetricsNotifier) DecrementRequestsInFlightCounter(requestType string, delta int) {
	atomic.AddUint64(&notifier.requestMetricForType(requestType).RequestsInFlight, uint64(-delta))
}

func (notifier *RequestMetricsNotifier) UpdateLatency(requestType string, dur time.Duration) {
	addr := &notifier.requestMetricForType(requestType).RequestLatencyMax
	for {
		val := atomic.LoadInt64(addr)
		newval := int64(dur)
		if newval < val {
			return
		}

		if atomic.CompareAndSwapInt64(addr, val, newval) {
			return
		}
	}
}

func (notifier *RequestMetricsNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.logger.Session("request-metrics-notifier")
	logger.Info("starting", lager.Data{"interval": notifier.metricsInterval})
	defer logger.Info("completed")
	close(ready)

	tick := notifier.clock.NewTicker(notifier.metricsInterval)

	for {
		select {
		case <-signals:
			return nil

		case <-tick.C():
			logger.Debug("emitting-metrics")

			var err error
			for _, requestType := range []string{"Lock", "Release", "Fetch", "FetchAll"} {
				metric := notifier.requestMetricForType(requestType)
				opt := loggregator.WithEnvelopeTag("request-type", requestType)

				err = notifier.metronClient.SendMetric(requestsStartedMetric, int(atomic.LoadUint64(&metric.RequestsStarted)), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-started-metric", err)
				}
				err = notifier.metronClient.SendMetric(requestsSucceededMetric, int(atomic.LoadUint64(&metric.RequestsSucceeded)), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-succeeded-metric", err)
				}
				err = notifier.metronClient.SendMetric(requestsFailedMetric, int(atomic.LoadUint64(&metric.RequestsFailed)), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-failed-metric", err)
				}
				err = notifier.metronClient.SendMetric(requestsInFlightMetric, int(atomic.LoadUint64(&metric.RequestsInFlight)), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-in-flight-metric", err)
				}
				latency := atomic.SwapInt64(&metric.RequestLatencyMax, 0)
				err = notifier.metronClient.SendDuration(requestLatencyMaxDuration, time.Duration(latency), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-latency-max-metric", err)
				}
			}
		}
	}
}
