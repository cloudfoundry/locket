package db

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
)

//go:generate counterfeiter . LockDB
type LockDB interface {
	Lock(logger lager.Logger, resource *models.Resource, ttl int64) (*Lock, error)
	Release(logger lager.Logger, resource *models.Resource) error
	Fetch(logger lager.Logger, key string) (*Lock, error)
	FetchAll(logger lager.Logger, lockType string) ([]*Lock, error)
}

type Lock struct {
	*models.Resource
	TtlInSeconds  int64
	ModifiedIndex int64
}

type SQLDB struct {
	db     *sql.DB
	flavor string
	helper helpers.SQLHelper
}

func NewSQLDB(
	db *sql.DB,
	flavor string,
) *SQLDB {
	helper := helpers.NewSQLHelper(flavor)
	return &SQLDB{
		db:     db,
		flavor: flavor,
		helper: helper,
	}
}
