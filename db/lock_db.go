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
			"key":   resource.Key,
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
		var owner string
		row := db.helper.One(logger, tx, "locks", []string{"owner"}, true, "key = ?", resource.Key)
		err := row.Scan(&owner)
		if err != nil {
			return err
		}

		if owner != resource.Owner {
			logger.Error("cannot-release-lock", models.ErrLockCollision, lager.Data{
				"key":   resource.Key,
				"owner": owner,
				"actor": resource.Owner,
			})

			return models.ErrLockCollision
		}

		_, err = db.helper.Delete(logger, tx, "locks", "key = ?", resource.Key)
		return err
	})
}

func (db *SQLDB) Fetch(logger lager.Logger, key string) (*models.Resource, error) {
	logger = logger.Session("release-lock")
	var resource *models.Resource

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		var value, owner string
		row := db.helper.One(logger, tx, "locks", []string{"value", "owner"}, true, "key = ?", key)
		err := row.Scan(&value, &owner)
		if err != nil {
			return err
		}

		resource = &models.Resource{Key: key, Value: value, Owner: owner}
		return nil
	})

	return resource, err
}
