package metrics

import (
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/clock"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/lager"
)

const (
	requestsStartedMetric     = "RequestsStarted"
	requestsSucceededMetric   = "RequestsSucceeded"
	requestsFailedMetric      = "RequestsFailed"
	requestsInFlightMetric    = "RequestsInFlight"
	requestLatencyMaxDuration = "RequestLatencyMax"
)

//go:generate counterfeiter . RequestMetrics
type RequestMetrics interface {
	IncrementRequestsStartedCounter(requestType string, delta int)
	IncrementRequestsSucceededCounter(requestType string, delta int)
	IncrementRequestsFailedCounter(requestType string, delta int)
	IncrementRequestsInFlightCounter(requestType string, delta int)
	DecrementRequestsInFlightCounter(requestType string, delta int)
	UpdateLatency(requestType string, dur time.Duration)

	RequestsStarted(requestType string) uint64
	RequestsSucceeded(requestType string) uint64
	RequestsFailed(requestType string) uint64
	RequestsInFlight(requestType string) uint64
	MaxLatency(requestType string) time.Duration
}

type RequestMetricsNotifier struct {
	logger          lager.Logger
	ticker          clock.Clock
	metricsInterval time.Duration
	metronClient    loggingclient.IngressClient

	requestsStarted       map[string]uint64
	requestsSucceeded     map[string]uint64
	requestsFailed        map[string]uint64
	requestsInFlight      map[string]uint64
	latencyMax            map[string]time.Duration
	latencyLock           *sync.Mutex
	requestsStartedLock   *sync.Mutex
	requestsSucceededLock *sync.Mutex
	requestsFailedLock    *sync.Mutex
	requestsInFlightLock  *sync.Mutex
}

func NewRequestMetricsNotifier(logger lager.Logger, ticker clock.Clock, metronClient loggingclient.IngressClient, metricsInterval time.Duration) *RequestMetricsNotifier {
	return &RequestMetricsNotifier{
		logger:                logger,
		ticker:                ticker,
		metricsInterval:       metricsInterval,
		metronClient:          metronClient,
		latencyLock:           &sync.Mutex{},
		requestsStartedLock:   &sync.Mutex{},
		requestsSucceededLock: &sync.Mutex{},
		requestsFailedLock:    &sync.Mutex{},
		requestsInFlightLock:  &sync.Mutex{},
		requestsStarted:       map[string]uint64{},
		requestsSucceeded:     map[string]uint64{},
		requestsFailed:        map[string]uint64{},
		requestsInFlight:      map[string]uint64{},
		latencyMax:            map[string]time.Duration{},
	}
}

func (notifier *RequestMetricsNotifier) IncrementRequestsStartedCounter(requestType string, delta int) {
	notifier.requestsStartedLock.Lock()
	defer notifier.requestsStartedLock.Unlock()

	notifier.requestsStarted[requestType] += uint64(delta)
}

func (notifier *RequestMetricsNotifier) IncrementRequestsSucceededCounter(requestType string, delta int) {
	notifier.requestsSucceededLock.Lock()
	defer notifier.requestsSucceededLock.Unlock()

	notifier.requestsSucceeded[requestType] += uint64(delta)
}

func (notifier *RequestMetricsNotifier) IncrementRequestsFailedCounter(requestType string, delta int) {
	notifier.requestsFailedLock.Lock()
	defer notifier.requestsFailedLock.Unlock()

	notifier.requestsFailed[requestType] += uint64(delta)
}

func (notifier *RequestMetricsNotifier) IncrementRequestsInFlightCounter(requestType string, delta int) {
	notifier.requestsInFlightLock.Lock()
	defer notifier.requestsInFlightLock.Unlock()

	notifier.requestsInFlight[requestType] += uint64(delta)
}

func (notifier *RequestMetricsNotifier) DecrementRequestsInFlightCounter(requestType string, delta int) {
	notifier.requestsInFlightLock.Lock()
	defer notifier.requestsInFlightLock.Unlock()

	notifier.requestsInFlight[requestType] -= uint64(delta)
}

func (notifier *RequestMetricsNotifier) UpdateLatency(requestType string, dur time.Duration) {
	notifier.latencyLock.Lock()
	defer notifier.latencyLock.Unlock()

	if dur > notifier.latencyMax[requestType] {
		notifier.latencyMax[requestType] = dur
	}
}

func (notifier *RequestMetricsNotifier) RequestsStarted(requestType string) uint64 {
	notifier.requestsStartedLock.Lock()
	defer notifier.requestsStartedLock.Unlock()

	return notifier.requestsStarted[requestType]
}

func (notifier *RequestMetricsNotifier) RequestsSucceeded(requestType string) uint64 {
	notifier.requestsSucceededLock.Lock()
	defer notifier.requestsSucceededLock.Unlock()

	return notifier.requestsSucceeded[requestType]
}

func (notifier *RequestMetricsNotifier) RequestsFailed(requestType string) uint64 {
	notifier.requestsFailedLock.Lock()
	defer notifier.requestsFailedLock.Unlock()

	return notifier.requestsFailed[requestType]
}

func (notifier *RequestMetricsNotifier) RequestsInFlight(requestType string) uint64 {
	notifier.requestsInFlightLock.Lock()
	defer notifier.requestsInFlightLock.Unlock()

	return notifier.requestsInFlight[requestType]
}

func (notifier *RequestMetricsNotifier) readAndResetMaxLatency(requestType string) time.Duration {
	notifier.latencyLock.Lock()
	defer notifier.latencyLock.Unlock()

	max := notifier.latencyMax[requestType]
	notifier.latencyMax[requestType] = 0

	return max
}

func (notifier *RequestMetricsNotifier) MaxLatency(requestType string) time.Duration {
	notifier.latencyLock.Lock()
	defer notifier.latencyLock.Unlock()

	return notifier.latencyMax[requestType]
}

func (notifier *RequestMetricsNotifier) requestsStartedTags() []string {
	notifier.requestsStartedLock.Lock()
	defer notifier.requestsStartedLock.Unlock()

	tags := []string{}

	for requestType, _ := range notifier.requestsStarted {
		tags = append(tags, requestType)
	}

	return tags
}

func (notifier *RequestMetricsNotifier) requestsSucceededTags() []string {
	notifier.requestsSucceededLock.Lock()
	defer notifier.requestsSucceededLock.Unlock()

	tags := []string{}

	for requestType, _ := range notifier.requestsSucceeded {
		tags = append(tags, requestType)
	}

	return tags
}

func (notifier *RequestMetricsNotifier) requestsFailedTags() []string {
	notifier.requestsFailedLock.Lock()
	defer notifier.requestsFailedLock.Unlock()

	tags := []string{}

	for requestType, _ := range notifier.requestsFailed {
		tags = append(tags, requestType)
	}

	return tags
}

func (notifier *RequestMetricsNotifier) requestsInFlightTags() []string {
	notifier.requestsInFlightLock.Lock()
	defer notifier.requestsInFlightLock.Unlock()

	tags := []string{}

	for requestType, _ := range notifier.requestsInFlight {
		tags = append(tags, requestType)
	}

	return tags
}

func (notifier *RequestMetricsNotifier) latencyMaxTags() []string {
	notifier.latencyLock.Lock()
	defer notifier.latencyLock.Unlock()

	tags := []string{}

	for requestType, _ := range notifier.latencyMax {
		tags = append(tags, requestType)
	}

	return tags
}

func (notifier *RequestMetricsNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.logger.Session("request-metrics-notifier")
	logger.Info("starting", lager.Data{"interval": notifier.metricsInterval})
	defer logger.Info("completed")
	close(ready)

	tick := notifier.ticker.NewTicker(notifier.metricsInterval)

	for {
		select {
		case <-signals:
			return nil
		case <-tick.C():
			logger.Debug("emitting-metrics")

			var err error
			for _, requestType := range notifier.requestsStartedTags() {
				opt := loggregator.WithEnvelopeTag("request-type", requestType)
				value := notifier.RequestsStarted(requestType)
				err = notifier.metronClient.SendMetric(requestsStartedMetric, int(value), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-started-metric", err)
				}
			}

			for _, requestType := range notifier.requestsSucceededTags() {
				opt := loggregator.WithEnvelopeTag("request-type", requestType)
				value := notifier.RequestsSucceeded(requestType)
				err = notifier.metronClient.SendMetric(requestsSucceededMetric, int(value), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-succeeded-metric", err)
				}
			}

			for _, requestType := range notifier.requestsFailedTags() {
				opt := loggregator.WithEnvelopeTag("request-type", requestType)
				value := notifier.RequestsFailed(requestType)
				err = notifier.metronClient.SendMetric(requestsFailedMetric, int(value), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-failed-metric", err)
				}
			}

			for _, requestType := range notifier.requestsInFlightTags() {
				opt := loggregator.WithEnvelopeTag("request-type", requestType)
				value := notifier.RequestsInFlight(requestType)
				err = notifier.metronClient.SendMetric(requestsInFlightMetric, int(value), opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-in-flight-metric", err)
				}
			}

			for _, requestType := range notifier.latencyMaxTags() {
				opt := loggregator.WithEnvelopeTag("request-type", requestType)
				max := notifier.readAndResetMaxLatency(requestType)
				err = notifier.metronClient.SendDuration(requestLatencyMaxDuration, max, opt)
				if err != nil {
					logger.Error("failed-to-emit-requests-latency-max-metric", err)
				}
			}
		}
	}

	return nil
}
