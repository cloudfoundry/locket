package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/locket/cmd/locket/config"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/expiration"
	"code.cloudfoundry.org/locket/grpcserver"
	"code.cloudfoundry.org/locket/handlers"
)

var configFilePath = flag.String(
	"config",
	"",
	"Path to Locket JSON Configuration file",
)

func main() {
	flag.Parse()

	var sqlDB *db.SQLDB

	cfg, err := config.NewLocketConfig(*configFilePath)
	if err != nil {
		panic(err)
	}

	logger, reconfigurableSink := lagerflags.NewFromConfig("locket", cfg.LagerConfig)
	clock := clock.NewClock()

	if cfg.DatabaseDriver != "" && cfg.DatabaseConnectionString != "" {
		var err error

		if cfg.DatabaseDriver == "postgres" && !strings.Contains(cfg.DatabaseConnectionString, "sslmode") {
			cfg.DatabaseConnectionString = fmt.Sprintf("%s?sslmode=disable", cfg.DatabaseConnectionString)
		}

		sqlConn, err := sql.Open(cfg.DatabaseDriver, cfg.DatabaseConnectionString)
		if err != nil {
			logger.Fatal("failed-to-open-sql", err)
		}
		defer sqlConn.Close()

		err = sqlConn.Ping()
		if err != nil {
			logger.Fatal("sql-failed-to-connect", err)
		}

		sqlDB = db.NewSQLDB(
			sqlConn,
			cfg.DatabaseDriver,
		)

		err = sqlDB.SetIsolationLevel(logger, helpers.IsolationLevelReadCommitted)
		if err != nil {
			logger.Fatal("sql-failed-to-set-isolation-level", err)
		}
	}

	if sqlDB == nil {
		logger.Fatal("no-database-configured", errors.New("no database configured"))
	}

	err = sqlDB.CreateLockTable(logger)
	if err != nil {
		logger.Fatal("failed-to-create-lock-table", err)
	}

	consulClient, err := consuladapter.NewClientFromUrl(cfg.ConsulCluster)
	if err != nil {
		logger.Fatal("new-consul-client-failed", err)
	}

	_, portString, err := net.SplitHostPort(cfg.ListenAddress)
	if err != nil {
		logger.Fatal("failed-invalid-listen-address", err)
	}
	portNum, err := net.LookupPort("tcp", portString)
	if err != nil {
		logger.Fatal("failed-invalid-listen-port", err)
	}

	lockPick := expiration.NewLockPick(sqlDB, clock)
	burglar := expiration.NewBurglar(logger, sqlDB, lockPick, clock, locket.RetryInterval)
	handler := handlers.NewLocketHandler(logger, sqlDB, lockPick)
	server := grpcserver.NewGRPCServer(logger, cfg.ListenAddress, handler)
	registrationRunner := initializeRegistrationRunner(logger, consulClient, portNum, clock)
	members := grouper.Members{
		{"server", server},
		{"burglar", burglar},
		{"registration-runner", registrationRunner},
	}

	if cfg.DebugAddress != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(cfg.DebugAddress, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)
	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}
}

func initializeRegistrationRunner(
	logger lager.Logger,
	consulClient consuladapter.Client,
	port int,
	clock clock.Clock,
) ifrit.Runner {
	registration := &api.AgentServiceRegistration{
		Name: "locket",
		Port: port,
		Check: &api.AgentServiceCheck{
			TTL: "20s",
		},
	}
	return locket.NewRegistrationRunner(logger, registration, consulClient, locket.RetryInterval, clock)
}
