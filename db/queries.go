package db

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
)

func (db *SQLDB) CreateLockTable(logger lager.Logger) error {
	var query string
	if db.flavor == helpers.MSSQL {
		query = `
			IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='locks' AND xtype='U')
			CREATE TABLE locks (
				path VARCHAR(255) PRIMARY KEY,
				owner VARCHAR(255),
				value VARCHAR(255),
				modified_index BIGINT DEFAULT 0,
				ttl BIGINT DEFAULT 0
			);
		`
	} else {
		query = `
			CREATE TABLE IF NOT EXISTS locks (
				path VARCHAR(255) PRIMARY KEY,
				owner VARCHAR(255),
				value VARCHAR(255),
				modified_index BIGINT DEFAULT 0,
				ttl BIGINT DEFAULT 0
			);
		`
	}
	_, err := db.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLDB) SetIsolationLevel(logger lager.Logger, level string) error {
	return db.helper.SetIsolationLevel(logger, db.db, level)
}
