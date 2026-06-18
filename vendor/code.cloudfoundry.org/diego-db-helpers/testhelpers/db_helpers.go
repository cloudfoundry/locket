package testhelpers

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/diego-db-helpers/testhelpers/sqlrunner"
)

const (
	mysqlFlavor    = "mysql"
	mysql8Flavor   = "mysql8"
	postgresFlavor = "postgres"
)

func driver() string {
	flavor := os.Getenv("DB")
	if flavor == "" {
		flavor = postgresFlavor
	}
	return flavor
}

func UseMySQL() bool {
	d := driver()
	return d == mysqlFlavor || d == mysql8Flavor
}

func UsePostgres() bool {
	return driver() == postgresFlavor
}

func NewSQLRunner(dbName string) sqlrunner.SQLRunner {
	if UseMySQL() {
		return sqlrunner.NewMySQLRunner(dbName)
	} else if UsePostgres() {
		return sqlrunner.NewPostgresRunner(dbName)
	}
	panic(fmt.Sprintf("driver '%s' is not supported", driver()))
}
