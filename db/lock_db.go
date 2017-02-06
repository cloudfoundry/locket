package db

import (
	"database/sql"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
)

func (db *SQLDB) Lock(logger lager.Logger, resource *models.Resource, ttl int64) error {
	logger = logger.Session("lock")

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		res, index, err := db.fetchLock(logger, tx, resource.Key)
		if err != nil {
			sqlErr := db.helper.ConvertSQLError(err)
			if sqlErr != helpers.ErrResourceNotFound {
				logger.Error("failed-to-fetch-lock", err)
				return err
			}
			logger.Info("lock-does-not-exist", lager.Data{"index": index})
		} else if res.Owner != resource.Owner {
			logger.Error("lock-already-exists", err)
			return models.ErrLockCollision
		}

		index++

		_, err = db.helper.Upsert(logger, tx, "locks",
			helpers.SQLAttributes{
				"path": resource.Key,
			},
			helpers.SQLAttributes{
				"owner":          resource.Owner,
				"value":          resource.Value,
				"modified_index": index,
			},
		)
		if err != nil {
			logger.Error("failed-updating-lock", err)
		} else {
			go db.checkExpiration(logger, resource, ttl, index)
		}

		return err
	})

	return err
}

func (db *SQLDB) fetchLock(logger lager.Logger, q helpers.Queryable, key string) (*models.Resource, int64, error) {
	row := db.helper.One(logger, q, "locks",
		helpers.ColumnList{"owner", "value", "modified_index"},
		helpers.LockRow,
		"path = ?", key,
	)

	var owner string
	var value string
	var index int64
	err := row.Scan(&owner, &value, &index)
	if err != nil {
		return nil, 0, err
	}

	if owner == "" {
		return nil, index, helpers.ErrResourceNotFound
	}

	return &models.Resource{
		Key:   key,
		Owner: owner,
		Value: value,
	}, index, nil
}

func (db *SQLDB) Release(logger lager.Logger, resource *models.Resource) error {
	logger = logger.Session("release-lock")

	return db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		res, index, err := db.fetchLock(logger, tx, resource.Key)
		if err != nil {
			logger.Error("failed-to-fetch-lock", err)
			return err
		}

		if res.Owner != resource.Owner {
			logger.Error("cannot-release-lock", models.ErrLockCollision)
			return models.ErrLockCollision
		}

		index++
		_, err = db.helper.Update(logger, tx, "locks",
			helpers.SQLAttributes{
				"value":          "",
				"owner":          "",
				"modified_index": index,
			},
			"path = ?", resource.Key,
		)
		if err != nil {
			logger.Error("failed-to-release-lock", err)
		}
		return err
	})
}

func (db *SQLDB) Fetch(logger lager.Logger, key string) (*models.Resource, error) {
	logger = logger.Session("release-lock")
	var resource *models.Resource

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		var err error

		resource, _, err = db.fetchLock(logger, tx, key)
		if err != nil {
			logger.Error("failed-to-fetch-lock", err)
			return err
		}

		return nil
	})

	return resource, err
}

func (db *SQLDB) checkExpiration(logger lager.Logger, res *models.Resource, ttl, index int64) {
	select {
	case <-db.clock.NewTimer(time.Duration(ttl) * time.Second).C():
		db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
			updatedIndex := index + 1

			result, err := db.helper.Update(logger, tx, "locks",
				helpers.SQLAttributes{
					"value":          "",
					"owner":          "",
					"modified_index": updatedIndex,
				},
				"path = ? AND modified_index = ?",
				res.Key, index,
			)

			if err != nil {
				logger.Error("failed-expiring-lock", err)
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				logger.Error("failed-to-get-rows-affected", err)
			}
			if rowsAffected > 0 {
				logger.Info("expired-lock", lager.Data{"lock": res, "ttl": ttl})
			}

			return err
		})
	}
}
