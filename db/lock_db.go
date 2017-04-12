package db

import (
	"database/sql"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
)

func lagerDataFromLock(resource *models.Resource) lager.Data {
	return lager.Data{
		"key":   resource.GetKey(),
		"owner": resource.GetOwner(),
		"type":  resource.GetType(),
	}
}

func (db *SQLDB) Lock(logger lager.Logger, resource *models.Resource, ttl int64) (*Lock, error) {
	logger = logger.Session("lock", lagerDataFromLock(resource))
	var lock *Lock

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		newLock := false

		res, index, _, err := db.fetchLock(logger, tx, resource.Key)
		if err != nil {
			sqlErr := db.helper.ConvertSQLError(err)
			if sqlErr != helpers.ErrResourceNotFound {
				logger.Error("failed-to-fetch-lock", err)
				return err
			}
			newLock = true
		} else if res.Owner != resource.Owner {
			logger.Debug("lock-already-exists")
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
				"type":           lock.Type,
				"modified_index": lock.ModifiedIndex,
				"ttl":            lock.TtlInSeconds,
			},
		)
		if err != nil {
			logger.Error("failed-updating-lock", err)
			return err
		}
		if newLock {
			logger.Info("acquired-lock")
		}
		return nil

	})

	return lock, err
}

func (db *SQLDB) Release(logger lager.Logger, resource *models.Resource) error {
	logger = logger.Session("release-lock", lagerDataFromLock(resource))

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
				"type":           "",
				"modified_index": index,
			},
			"path = ?", resource.Key,
		)
		if err != nil {
			logger.Error("failed-to-release-lock", err)
			return err
		}
		logger.Info("released-lock")
		return nil
	})
}

func (db *SQLDB) Fetch(logger lager.Logger, key string) (*Lock, error) {
	logger = logger.Session("fetch-lock", lager.Data{"key": key})
	var lock *Lock

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		res, index, ttl, err := db.fetchLock(logger, tx, key)
		if err != nil {
			logger.Error("failed-to-fetch-lock", err)
			sqlErr := db.helper.ConvertSQLError(err)
			if sqlErr == helpers.ErrResourceNotFound {
				return models.ErrResourceNotFound
			}
			return sqlErr
		}

		lock = &Lock{Resource: res, ModifiedIndex: index, TtlInSeconds: ttl}

		return nil
	})

	return lock, err
}

func (db *SQLDB) FetchAll(logger lager.Logger, lockType string) ([]*Lock, error) {
	logger = logger.Session("fetch-all-locks", lager.Data{"type": lockType})
	var locks []*Lock

	err := db.helper.Transact(logger, db.db, func(logger lager.Logger, tx *sql.Tx) error {
		var where string
		whereBindings := make([]interface{}, 0)

		if lockType != "" {
			where = "type = ?"
			whereBindings = append(whereBindings, lockType)
		}

		rows, err := db.helper.All(logger, tx, "locks",
			helpers.ColumnList{"path", "owner", "value", "type", "modified_index", "ttl"},
			helpers.LockRow, where, whereBindings...,
		)
		if err != nil {
			logger.Error("failed-to-fetch-locks", err)
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var key, owner, value, lockType string
			var index, ttl int64

			err := rows.Scan(&key, &owner, &value, &lockType, &index, &ttl)
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
					Type:  lockType,
				},
				ModifiedIndex: index,
				TtlInSeconds:  ttl,
			})
		}

		return nil
	})

	return locks, err
}

func (db *SQLDB) Count(logger lager.Logger, lockType string) (int, error) {
	whereBindings := make([]interface{}, 0)
	wheres := "owner <> ?"
	whereBindings = append(whereBindings, "")

	if lockType != "" {
		wheres += " AND type = ?"
		whereBindings = append(whereBindings, lockType)
	}

	logger = logger.Session("count-locks")
	return db.helper.Count(logger, db.db, "locks", wheres, whereBindings...)
}

func (db *SQLDB) fetchLock(logger lager.Logger, q helpers.Queryable, key string) (*models.Resource, int64, int64, error) {
	row := db.helper.One(logger, q, "locks",
		helpers.ColumnList{"owner", "value", "type", "modified_index", "ttl"},
		helpers.LockRow,
		"path = ?", key,
	)

	var owner, value, lockType string
	var index, ttl int64
	err := row.Scan(&owner, &value, &lockType, &index, &ttl)
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
		Type:  lockType,
	}, index, ttl, nil
}
