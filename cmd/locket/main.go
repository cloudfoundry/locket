package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/locket/cmd/locket/config"
	"code.cloudfoundry.org/locket/db"
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

	logger, _ := lagerflags.NewFromConfig("locket", cfg.LagerConfig)

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

	handler := handlers.NewLocketHandler(logger, sqlDB)
	server := grpcserver.NewGRPCServer(logger, cfg.ListenAddress, handler)
	members := grouper.Members{
		{"server", server},
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
