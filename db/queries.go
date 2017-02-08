package db

import "code.cloudfoundry.org/lager"

func (db *SQLDB) CreateLockTable(logger lager.Logger) error {
	_, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS locks (
			path VARCHAR(255) PRIMARY KEY,
			owner VARCHAR(255),
			value VARCHAR(255),
			modified_index BIGINT DEFAULT 0,
			ttl BIGINT DEFAULT 0
		);
	`)
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLDB) SetIsolationLevel(logger lager.Logger, level string) error {
	return db.helper.SetIsolationLevel(logger, db.db, level)
}
