package expiration

import (
	"sync"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/db"
)

//go:generate counterfeiter . LockPick
type LockPick interface {
	RegisterTTL(logger lager.Logger, lock *db.Lock)
}

type lockPick struct {
	lockDB    db.LockDB
	clock     clock.Clock
	lockTTLs  map[ttlKey]struct{}
	lockMutex *sync.Mutex
}

type ttlKey struct {
	key   string
	index int64
}

func NewLockPick(lockDB db.LockDB, clock clock.Clock) lockPick {
	return lockPick{
		lockDB:    lockDB,
		clock:     clock,
		lockTTLs:  make(map[ttlKey]struct{}),
		lockMutex: &sync.Mutex{},
	}
}

func (l lockPick) RegisterTTL(logger lager.Logger, lock *db.Lock) {
	logger = logger.Session("register-ttl", lager.Data{"key": lock.Key, "modified-index": lock.ModifiedIndex})
	logger.Debug("starting")
	logger.Debug("completed")

	l.lockMutex.Lock()
	defer l.lockMutex.Unlock()

	key := ttlKey{
		lock.Key,
		lock.ModifiedIndex,
	}

	_, ok := l.lockTTLs[key]
	if !ok {
		go l.checkExpiration(logger, lock)
		l.lockTTLs[key] = struct{}{}
	}
}

func (l lockPick) checkExpiration(logger lager.Logger, lock *db.Lock) {
	defer func() {
		key := ttlKey{
			lock.Key,
			lock.ModifiedIndex,
		}

		l.lockMutex.Lock()
		delete(l.lockTTLs, key)
		l.lockMutex.Unlock()
	}()

	select {
	case <-l.clock.NewTimer(time.Duration(lock.TtlInSeconds) * time.Second).C():
		fetchedLock, err := l.lockDB.Fetch(logger, lock.Key)
		if err != nil {
			return
		}

		if fetchedLock.ModifiedIndex == lock.ModifiedIndex {
			logger.Info("lock-expired")
			err = l.lockDB.Release(logger, lock.Resource)
			if err != nil {
				return
			}
		}
	}
}
