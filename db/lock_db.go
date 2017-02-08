package db

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
)

func (db *SQLDB) Lock(logger lager.Logger, resource *models.Resource, ttl int64) (*Lock, error) {
	logger = logger.Session("lock")
	var lock *Lock

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		res, index, _, err := db.fetchLock(logger, tx, resource.Key)
		if err != nil {
			sqlErr := db.helper.ConvertSQLError(err)
			if sqlErr != helpers.ErrResourceNotFound {
				logger.Error("failed-to-fetch-lock", err)
				return err
			}
			logger.Info("lock-does-not-exist")
		} else if res.Owner != resource.Owner {
			logger.Error("lock-already-exists", err)
			return models.ErrLockCollision
		}

		index++

		lock = &Lock{
			Resource:      resource,
			ModifiedIndex: index,
			TtlInSeconds:  ttl,
		}

		_, err = db.helper.Upsert(logger, tx, "locks",
			helpers.SQLAttributes{
				"path": lock.Key,
			},
			helpers.SQLAttributes{
				"owner":          lock.Owner,
				"value":          lock.Value,
				"modified_index": lock.ModifiedIndex,
				"ttl":            lock.TtlInSeconds,
			},
		)
		if err != nil {
			logger.Error("failed-updating-lock", err)
		}

		return err
	})

	return lock, err
}

func (db *SQLDB) Release(logger lager.Logger, resource *models.Resource) error {
	logger = logger.Session("release-lock")

	return db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		res, index, _, err := db.fetchLock(logger, tx, resource.Key)
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

func (db *SQLDB) Fetch(logger lager.Logger, key string) (*Lock, error) {
	logger = logger.Session("fetch-lock")
	var lock *Lock

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		res, index, ttl, err := db.fetchLock(logger, tx, key)
		if err != nil {
			logger.Error("failed-to-fetch-lock", err)
			return err
		}

		lock = &Lock{Resource: res, ModifiedIndex: index, TtlInSeconds: ttl}

		return nil
	})

	return lock, err
}

func (db *SQLDB) FetchAll(logger lager.Logger) ([]*Lock, error) {
	logger = logger.Session("fetch-all-locks")
	var locks []*Lock

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		rows, err := db.helper.All(logger, tx, "locks",
			helpers.ColumnList{"path", "owner", "value", "modified_index", "ttl"},
			helpers.LockRow, "",
		)
		if err != nil {
			logger.Error("failed-to-fetch-locks", err)
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var key, owner, value string
			var index, ttl int64

			err := rows.Scan(&key, &owner, &value, &index, &ttl)
			if err != nil {
				logger.Error("failed-to-scan-lock", err)
				continue
			}

			if owner == "" {
				continue
			}

			locks = append(locks, &Lock{
				Resource: &models.Resource{
					Key:   key,
					Owner: owner,
					Value: value,
				},
				ModifiedIndex: index,
				TtlInSeconds:  ttl,
			})
		}

		return nil
	})

	return locks, err
}

func (db *SQLDB) fetchLock(logger lager.Logger, q helpers.Queryable, key string) (*models.Resource, int64, int64, error) {
	row := db.helper.One(logger, q, "locks",
		helpers.ColumnList{"owner", "value", "modified_index", "ttl"},
		helpers.LockRow,
		"path = ?", key,
	)

	var owner string
	var value string
	var index, ttl int64
	err := row.Scan(&owner, &value, &index, &ttl)
	if err != nil {
		return nil, 0, 0, err
	}

	if owner == "" {
		return nil, index, 0, helpers.ErrResourceNotFound
	}

	return &models.Resource{
		Key:   key,
		Owner: owner,
		Value: value,
	}, index, ttl, nil
}
