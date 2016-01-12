package locket

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

var (
	ErrLockLost = errors.New("lock lost")
)

type Lock struct {
	consul *consuladapter.Session
	key    string
	value  []byte

	clock         clock.Clock
	retryInterval time.Duration

	logger lager.Logger

	lockAcquiredMetric metric.Metric
	lockUptimeMetric   metric.Duration
	lockAcquiredTime   time.Time
}

func NewLock(
	consul *consuladapter.Session,
	lockKey string,
	lockValue []byte,
	clock clock.Clock,
	retryInterval time.Duration,
	logger lager.Logger,
) Lock {
	lockMetricName := strings.Replace(lockKey, "/", "-", -1)
	return Lock{
		consul: consul,
		key:    lockKey,
		value:  lockValue,

		clock:         clock,
		retryInterval: retryInterval,

		logger: logger,

		lockAcquiredMetric: metric.Metric("LockHeld." + lockMetricName),
		lockUptimeMetric:   metric.Duration("LockHeldDuration." + lockMetricName),
	}
}

func (l Lock) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := l.logger.Session("lock", lager.Data{"key": l.key, "value": string(l.value)})
	logger.Info("starting")

	defer func() {
		l.consul.Destroy()
		logger.Info("done")
	}()

	acquireErr := make(chan error, 1)

	acquire := func(session *consuladapter.Session) {
		logger.Info("acquiring-lock")
		acquireErr <- session.AcquireLock(l.key, l.value)
	}

	var c <-chan time.Time
	var reemit <-chan time.Time

	go acquire(l.consul)

	for {
		select {
		case sig := <-signals:
			logger.Info("shutting-down", lager.Data{"received-signal": sig})

			logger.Debug("releasing-lock")
			l.emitMetrics(false)
			return nil
		case err := <-l.consul.Err():
			var data lager.Data
			if err != nil {
				data = lager.Data{"err": err.Error()}
			}

			if ready == nil {
				logger.Info("lost-lock", data)
				l.emitMetrics(false)
				return ErrLockLost
			}

			logger.Info("consul-error", data)
			c = l.clock.NewTimer(l.retryInterval).C()
		case err := <-acquireErr:
			if err != nil {
				logger.Info("acquire-lock-failed", lager.Data{"err": err.Error()})
				l.emitMetrics(false)
				c = l.clock.NewTimer(l.retryInterval).C()
				break
			}

			logger.Info("acquire-lock-succeeded")
			l.lockAcquiredTime = l.clock.Now()
			l.emitMetrics(true)
			reemit = l.clock.NewTimer(30 * time.Second).C()
			close(ready)
			ready = nil
			c = nil
			logger.Info("started")
		case <-reemit:
			l.emitMetrics(true)
			reemit = l.clock.NewTimer(30 * time.Second).C()
		case <-c:
			logger.Info("retrying-acquiring-lock")
			newSession, err := l.consul.Recreate()
			if err != nil {
				c = l.clock.NewTimer(l.retryInterval).C()
			} else {
				l.consul = newSession
				c = nil
				go acquire(newSession)
			}
		}
	}
}

func (l Lock) emitMetrics(acquired bool) {
	var acqVal int
	var uptime time.Duration

	if acquired {
		acqVal = 1
		uptime = l.clock.Since(l.lockAcquiredTime)
	} else {
		acqVal = 0
		uptime = 0
	}

	l.logger.Debug("reemit-lock-uptime", lager.Data{"uptime": uptime,
		"uptimeMetricName":       l.lockUptimeMetric,
		"lockAcquiredMetricName": l.lockAcquiredMetric,
	})
	err := l.lockUptimeMetric.Send(uptime)
	if err != nil {
		l.logger.Error("failed-to-send-lock-uptime-metric", err)
	}

	err = l.lockAcquiredMetric.Send(acqVal)
	if err != nil {
		l.logger.Error("failed-to-send-lock-acquired-metric", err)
	}
}
