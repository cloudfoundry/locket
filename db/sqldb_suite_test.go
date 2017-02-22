package db_test

import (
	"database/sql"
	"fmt"
	"os"
	"net"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/bbs/test_helpers/tool_helpers"
	"code.cloudfoundry.org/lager/lagertest"
	sqldb "code.cloudfoundry.org/locket/db"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/lib/pq"

	"testing"
)

var (
	rawDB                                *sql.DB
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
	} else if test_helpers.UseMSSQL() {
		dbDriverName = "mssql"
		dbBaseConnectionString = os.Getenv("MSSQL_BASE_CONNECTION_STRING")
		dbFlavor = helpers.MSSQL
	} else {
		panic("Unsupported driver")
	}

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	rawDB, err = sql.Open(dbDriverName, dbBaseConnectionString)
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())

	_, err = rawDB.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	if dbFlavor == helpers.MSSQL {
		_, err = rawDB.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))

		err = tool_helpers.Retry(5, func() error {
			var err error
			rawDB, err = sql.Open(dbDriverName, fmt.Sprintf("%s;database=diego_%d", dbBaseConnectionString, GinkgoParallelNode()))
			err = rawDB.Ping()
			return err
		})
		Expect(err).NotTo(HaveOccurred())

	} else {
		_, err = rawDB.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
		Expect(err).NotTo(HaveOccurred())

		rawDB, err = sql.Open(dbDriverName, fmt.Sprintf("%sdiego_%d", dbBaseConnectionString, GinkgoParallelNode()))
		Expect(err).NotTo(HaveOccurred())
		Expect(rawDB.Ping()).NotTo(HaveOccurred())
	}
	sqlDB = sqldb.NewSQLDB(rawDB, dbFlavor)
	err = sqlDB.CreateLockTable(logger)
	Expect(err).NotTo(HaveOccurred())

	// ensures sqlDB matches the db.DB interface
	var _ sqldb.LockDB = sqlDB
})

var _ = BeforeEach(func() {

	// ensure that all sqldb functions being tested only require one connection
	// to operate, otherwise a deadlock can be caused in bbs. For more
	// information see https://www.pivotaltracker.com/story/show/136754083
	rawDB.SetMaxOpenConns(1)
})

var _ = AfterEach(func() {
	truncateTables(rawDB)
})

var _ = AfterSuite(func() {
	Expect(rawDB.Close()).NotTo(HaveOccurred())
	rawDB, err := sql.Open(dbDriverName, dbBaseConnectionString)
	Expect(err).NotTo(HaveOccurred())
	Expect(rawDB.Ping()).NotTo(HaveOccurred())
	if dbFlavor == helpers.MSSQL {
		_, err = rawDB.Exec(fmt.Sprintf("IF EXISTS (SELECT * FROM master.dbo.sysdatabases WHERE name='diego_%d') DROP DATABASE diego_%d;", GinkgoParallelNode(), GinkgoParallelNode()))
		switch err.(type) {
		case *net.OpError:
			// On Azure, it may return a "i/o timeout" error when the database is dropped.
			// do nothing here
		default:
			Expect(err).NotTo(HaveOccurred())
		}
	} else {
		_, err = rawDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS diego_%d", GinkgoParallelNode()))
		Expect(err).NotTo(HaveOccurred())
	}
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
