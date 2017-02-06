package db

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
)

//go:generate counterfeiter . LockDB
type LockDB interface {
	Lock(logger lager.Logger, resource *models.Resource, ttl int64) error
	Release(logger lager.Logger, resource *models.Resource) error
	Fetch(logger lager.Logger, key string) (*models.Resource, error)
}

type SQLDB struct {
	db     *sql.DB
	flavor string
	helper helpers.SQLHelper
	clock  clock.Clock
}

func NewSQLDB(
	db *sql.DB,
	flavor string,
	clock clock.Clock,
) *SQLDB {
	helper := helpers.NewSQLHelper(flavor)
	return &SQLDB{
		db:     db,
		flavor: flavor,
		helper: helper,
		clock:  clock,
	}
}
