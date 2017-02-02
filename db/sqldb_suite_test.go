package db_test

import (
	"database/sql"
	"fmt"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/lager/lagertest"
	sqldb "code.cloudfoundry.org/locket/db"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/lib/pq"

	"testing"
)

var (
	db                                   *sql.DB
	sqlDB                                *sqldb.SQLDB
	logger                               *lagertest.TestLogger
	dbDriverName, dbBaseConnectionString string
	dbFlavor                             string
)

func TestSql(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "SQL DB Suite")
}

var _ = BeforeSuite(func() {
	var err error
	logger = lagertest.NewTestLogger("sql-db")

	if test_helpers.UsePostgres() {
		dbDriverName = "postgres"
		dbBaseConnectionString = "postgres://diego:diego_pw@localhost/"
		dbFlavor = helpers.Postgres
	} else if test_helpers.UseMySQL() {
		dbDriverName = "mysql"
		dbBaseConnectionString = "diego:diego_password@/"
		dbFlavor = helpers.MySQL
	} else {
		panic("Unsupported driver")
	}

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	db, err = sql.Open(dbDriverName, dbBaseConnectionString)
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	db, err = sql.Open(dbDriverName, fmt.Sprintf("%sdiego_%d", dbBaseConnectionString, GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	sqlDB = sqldb.NewSQLDB(db, dbFlavor)
	err = sqlDB.CreateLockTable(logger)
	Expect(err).NotTo(HaveOccurred())

	// ensures sqlDB matches the db.DB interface
	var _ sqldb.LockDB = sqlDB
})

var _ = BeforeEach(func() {

	// ensure that all sqldb functions being tested only require one connection
	// to operate, otherwise a deadlock can be caused in bbs. For more
	// information see https://www.pivotaltracker.com/story/show/136754083
	db.SetMaxOpenConns(1)
})

var _ = AfterEach(func() {
	truncateTables(db)
})

var _ = AfterSuite(func() {
	Expect(db.Close()).NotTo(HaveOccurred())
	db, err := sql.Open(dbDriverName, dbBaseConnectionString)
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Close()).NotTo(HaveOccurred())
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
