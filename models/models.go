package models

import "errors"

//go:generate bash ../scripts/generate_protos.sh
//go:generate counterfeiter . LocketClient

var ErrLockCollision = errors.New("lock-collision")
var ErrInvalidTTL = errors.New("invalid-ttl")
