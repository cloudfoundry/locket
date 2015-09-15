package locket

import (
	"path"
	"time"
)

const LockTTL = 10 * time.Second
const RetryInterval = 5 * time.Second

const LockSchemaRoot = "v1/locks"
const CellSchemaRoot = LockSchemaRoot + "/cell"
const ReceptorSchemaRoot = LockSchemaRoot + "/receptor"

func LockSchemaPath(lockName string) string {
	return path.Join(LockSchemaRoot, lockName)
}

func CellSchemaPath(cellID string) string {
	return path.Join(CellSchemaRoot, cellID)
}
