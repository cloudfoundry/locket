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
	activeLocksMetric        = "ActiveLocks"
	activePresencesMetric    = "ActivePresences"
	dbOpenConnectionsMetric  = "DBOpenConnections"
	dbQueriesTotalMetric     = "DBQueriesTotal"
	dbQueriesSucceededMetric = "DBQueriesSucceeded"
	dbQueriesFailedMetric    = "DBQueriesFailed"
	dbQueriesInFlightMetric  = "DBQueriesInFlight"
	dbQueryDurationMaxMetric = "DBQueryDurationMax"
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
			err := notifier.metronClient.SendMetric(dbOpenConnectionsMetric, notifier.lockDB.OpenConnections())
			if err != nil {
				logger.Error("failed-sending-db-open-connections-count", err)
			}

			err = notifier.metronClient.SendMetric(dbQueriesInFlightMetric, int(notifier.queryMonitor.QueriesInFlight()))
			if err != nil {
				logger.Error("inFlight-sending-db-queries-in-flight-count", err)
			}

			queriesTotal := notifier.queryMonitor.QueriesTotal()
			queriesSucceeded := notifier.queryMonitor.QueriesSucceeded()
			queriesFailed := notifier.queryMonitor.QueriesFailed()
			queryDurationMax := notifier.queryMonitor.ReadAndResetQueryDurationMax()

			logger.Debug("sending-queries-total-metric", lager.Data{"value": queriesTotal})
			err = notifier.metronClient.SendMetric(dbQueriesTotalMetric, int(queriesTotal))
			if err != nil {
				logger.Error("failed-sending-db-queries-total-count", err)
			}

			err = notifier.metronClient.SendMetric(dbQueriesSucceededMetric, int(queriesSucceeded))
			if err != nil {
				logger.Error("failed-sending-db-queries-succeeded-count", err)
			}

			err = notifier.metronClient.SendMetric(dbQueriesFailedMetric, int(queriesFailed))
			if err != nil {
				logger.Error("failed-sending-db-queries-failed-count", err)
			}

			err = notifier.metronClient.SendDuration(dbQueryDurationMaxMetric, queryDurationMax)
			if err != nil {
				logger.Error("inFlight-sending-db-query-duration-max", err)
			}

			locks, err := notifier.lockDB.Count(logger, models.LockType)
			if err != nil {
				logger.Error("failed-to-retrieve-lock-count", err)
			} else {
				err = notifier.metronClient.SendMetric(activeLocksMetric, locks)
				if err != nil {
					logger.Error("failed-sending-lock-count", err)
				}
			}

			presences, err := notifier.lockDB.Count(logger, models.PresenceType)
			if err != nil {
				logger.Error("failed-to-retrieve-presence-count", err)
			} else {
				err = notifier.metronClient.SendMetric(activePresencesMetric, presences)
				if err != nil {
					logger.Error("failed-sending-presences-count", err)
				}
			}

			logger.Debug("emitted-metrics")
		}
	}
	return nil
}
