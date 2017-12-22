package metrics

import (
	"os"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/clock"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/lager"
	"github.com/tedsuo/ifrit"
)

const (
	dbOpenConnectionsMetric  = "DBOpenConnections"
	dbQueriesTotalMetric     = "DBQueriesTotal"
	dbQueriesSucceededMetric = "DBQueriesSucceeded"
	dbQueriesFailedMetric    = "DBQueriesFailed"
	dbQueriesInFlightMetric  = "DBQueriesInFlight"
	dbQueryDurationMaxMetric = "DBQueryDurationMax"
)

type dbMetricsNotifier struct {
	logger          lager.Logger
	ticker          clock.Clock
	metricsInterval time.Duration
	lockDB          helpers.QueryableDB
	metronClient    loggingclient.IngressClient
	queryMonitor    helpers.QueryMonitor
}

func NewDBMetricsNotifier(logger lager.Logger, ticker clock.Clock, metronClient loggingclient.IngressClient, metricsInterval time.Duration, lockDB helpers.QueryableDB, queryMonitor helpers.QueryMonitor) ifrit.Runner {
	return &dbMetricsNotifier{
		logger:          logger,
		ticker:          ticker,
		metricsInterval: metricsInterval,
		lockDB:          lockDB,
		metronClient:    metronClient,
		queryMonitor:    queryMonitor,
	}
}

func (notifier *dbMetricsNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
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

			logger.Debug("emitted-metrics")
		}
	}
	return nil
}
