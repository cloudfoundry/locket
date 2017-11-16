package metrics

import (
	"os"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/clock"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
	"github.com/tedsuo/ifrit"
)

const (
	activeLocks        = "ActiveLocks"
	activePresences    = "ActivePresences"
	dbOpenConnections  = "DBOpenConnections"
	dbQueriesStarted   = "DBQueriesStarted"
	dbQueriesSucceeded = "DBQueriesSucceeded"
	dbQueriesFailed    = "DBQueriesFailed"
	dbQueriesInFlight  = "DBQueriesInFlight"
	dbQueryDurationMax = "DBQueryDurationMax"
)

type metricsNotifier struct {
	logger          lager.Logger
	ticker          clock.Clock
	metricsInterval time.Duration
	lockDB          LockDBMetrics
	metronClient    loggingclient.IngressClient
	queryMonitor    helpers.QueryMonitor
}

func NewMetricsNotifier(logger lager.Logger, ticker clock.Clock, metronClient loggingclient.IngressClient, metricsInterval time.Duration, lockDB LockDBMetrics, queryMonitor helpers.QueryMonitor) ifrit.Runner {
	return &metricsNotifier{
		logger:          logger,
		ticker:          ticker,
		metricsInterval: metricsInterval,
		lockDB:          lockDB,
		metronClient:    metronClient,
		queryMonitor:    queryMonitor,
	}
}

func (notifier *metricsNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := notifier.logger.Session("metrics-notifier")
	logger.Info("starting")
	defer logger.Info("compeleted")
	close(ready)

	tick := notifier.ticker.NewTicker(notifier.metricsInterval)
	for {
		select {
		case <-signals:
			return nil
		case <-tick.C():
			locks, err := notifier.lockDB.Count(logger, models.LockType)
			if err != nil {
				logger.Error("failed-to-retrieve-lock-count", err)
				continue
			}

			presences, err := notifier.lockDB.Count(logger, models.PresenceType)
			if err != nil {
				logger.Error("failed-to-retrieve-presence-count", err)
				continue
			}

			err = notifier.metronClient.SendMetric(activeLocks, locks)
			if err != nil {
				logger.Error("failed-sending-lock-count", err)
			}

			err = notifier.metronClient.SendMetric(activePresences, presences)
			if err != nil {
				logger.Error("failed-sending-presences-count", err)
			}

			err = notifier.metronClient.SendMetric(dbOpenConnections, notifier.lockDB.OpenConnections())
			if err != nil {
				logger.Error("failed-sending-db-open-connections-count", err)
			}

			err = notifier.metronClient.SendMetric(dbQueriesInFlight, int(notifier.queryMonitor.QueriesInFlight()))
			if err != nil {
				logger.Error("inFlight-sending-db-queries-in-flight-count", err)
			}

			queriesStarted := notifier.queryMonitor.QueriesStarted()
			queriesSucceeded := notifier.queryMonitor.QueriesSucceeded()
			queriesFailed := notifier.queryMonitor.QueriesFailed()
			queryDurationMax := notifier.queryMonitor.ReadAndResetQueryDurationMax()

			err = notifier.metronClient.SendMetric(dbQueriesStarted, int(queriesStarted))
			if err != nil {
				logger.Error("failed-sending-db-queries-started-count", err)
			}

			err = notifier.metronClient.SendMetric(dbQueriesSucceeded, int(queriesSucceeded))
			if err != nil {
				logger.Error("failed-sending-db-queries-succeeded-count", err)
			}

			err = notifier.metronClient.SendMetric(dbQueriesFailed, int(queriesFailed))
			if err != nil {
				logger.Error("failed-sending-db-queries-failed-count", err)
			}

			err = notifier.metronClient.SendDuration(dbQueryDurationMax, queryDurationMax)
			if err != nil {
				logger.Error("inFlight-sending-db-query-duration-max", err)
			}
		}
	}
	return nil
}
