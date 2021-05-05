package db_test

import (
	"context"
	"database/sql"
	"fmt"

	"code.cloudfoundry.org/diegosqldb"
	"code.cloudfoundry.org/diegosqldb/monitor"
	"code.cloudfoundry.org/diegosqldb/test_helpers"
	"code.cloudfoundry.org/lager/lagertest"
	sqldb "code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/guidprovider/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	rawDB                                *sql.DB
	sqlDB                                *sqldb.SQLDB
	logger                               *lagertest.TestLogger
	ctx                                  context.Context
	fakeGUIDProvider                     *fakes.FakeGUIDProvider
	dbDriverName, dbBaseConnectionString string
	dbFlavor                             string
	sqlHelper                            diegosqldb.SQLHelper
)

func TestSql(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "SQL DB Suite")
}

var _ = BeforeSuite(func() {
	var err error
	logger = lagertest.NewTestLogger("sql-db")
	ctx = context.Background()

	if test_helpers.UsePostgres() {
		dbDriverName = "postgres"
		dbBaseConnectionString = "postgres://diego:diego_pw@localhost/"
		dbFlavor = diegosqldb.Postgres
	} else if test_helpers.UseMySQL() {
		dbDriverName = "mysql"
		dbBaseConnectionString = "diego:diego_password@/"
		dbFlavor = diegosqldb.MySQL
	} else {
		panic("Unsupported driver")
	}

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	rawDB, err = diegosqldb.Connect(logger, dbDriverName, dbBaseConnectionString, "", false)

	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())

	_, err = rawDB.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	_, err = rawDB.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	connStringWithDB := fmt.Sprintf("%sdiego_%d", dbBaseConnectionString, GinkgoParallelNode())
	rawDB, err = diegosqldb.Connect(logger, dbDriverName, connStringWithDB, "", false)
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())

	fakeGUIDProvider = &fakes.FakeGUIDProvider{}
	db := diegosqldb.NewMonitoredDB(rawDB, monitor.New())
	sqlDB = sqldb.NewSQLDB(db, dbFlavor, fakeGUIDProvider)
	err = sqlDB.CreateLockTable(ctx, logger)
	Expect(err).NotTo(HaveOccurred())

	sqlHelper = diegosqldb.NewSQLHelper(dbFlavor)

	// ensures sqlDB matches the db.DB interface
	var _ sqldb.LockDB = sqlDB
})

var _ = BeforeEach(func() {

	// ensure that all sqldb functions being tested only require one connection
	// to operate, otherwise a deadlock can be caused in diegosqldb. For more
	// information see https://www.pivotaltracker.com/story/show/136754083
	rawDB.SetMaxOpenConns(1)
})

var _ = AfterEach(func() {
	truncateTables(rawDB)
})

var _ = AfterSuite(func() {
	Expect(rawDB.Close()).NotTo(HaveOccurred())
	rawDB, err := diegosqldb.Connect(logger, dbDriverName, dbBaseConnectionString, "", false)
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())
	_, err = rawDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Close()).NotTo(HaveOccurred())
})

func truncateTables(db *sql.DB) {
	for _, query := range truncateTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

var truncateTablesSQL = []string{
	"TRUNCATE TABLE locks",
}
