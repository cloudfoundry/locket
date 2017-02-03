package db

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
)

func (db *SQLDB) Lock(logger lager.Logger, resource *models.Resource) error {
	logger = logger.Session("lock")

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		_, err := db.helper.Insert(logger, tx, "locks", helpers.SQLAttributes{
			"path":  resource.Key,
			"owner": resource.Owner,
			"value": resource.Value,
		})

		modelErr := db.helper.ConvertSQLError(err)
		if modelErr == helpers.ErrResourceExists {
			logger.Error("lock-already-exists", err)
			return models.ErrLockCollision
		}

		return err
	})

	return err
}

func (db *SQLDB) Release(logger lager.Logger, resource *models.Resource) error {
	logger = logger.Session("release-lock")

	return db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		result, err := db.helper.Delete(logger, tx, "locks", "path = ? AND owner = ?", resource.Key, resource.Owner)

		n, err := result.RowsAffected()
		if err != nil {
			logger.Error("failed-to-get-rows-affected", err)
			return err
		}

		if n < 1 {
			logger.Error("cannot-release-lock", models.ErrLockCollision, lager.Data{
				"key":   resource.Key,
				"owner": resource.Owner,
				"actor": resource.Owner,
			})

			return models.ErrLockCollision
		}

		return err
	})
}

func (db *SQLDB) Fetch(logger lager.Logger, key string) (*models.Resource, error) {
	logger = logger.Session("release-lock")
	var resource *models.Resource

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		var value, owner string
		row := db.helper.One(logger, tx, "locks", []string{"value", "owner"}, true, "path = ?", key)
		err := row.Scan(&value, &owner)
		if err != nil {
			return err
		}

		resource = &models.Resource{Key: key, Value: value, Owner: owner}
		return nil
	})

	return resource, err
}
