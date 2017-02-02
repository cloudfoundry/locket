package lock

import (
	"context"
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
)

type lockRunner struct {
	logger lager.Logger

	locker        models.LocketClient
	lock          *models.Resource
	clock         clock.Clock
	retryInterval time.Duration
}

func NewLockRunner(
	logger lager.Logger,
	locker models.LocketClient,
	lock *models.Resource,
	clock clock.Clock,
	retryInterval time.Duration,
) *lockRunner {
	return &lockRunner{
		logger:        logger,
		locker:        locker,
		lock:          lock,
		clock:         clock,
		retryInterval: retryInterval,
	}
}

func (l *lockRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := l.logger.Session("lock")

	logger.Info("started")
	defer logger.Info("completed")

	close(ready)

	retry := l.clock.NewTimer(l.retryInterval)

	_, err := l.locker.Lock(context.Background(), &models.LockRequest{Resource: l.lock})
	if err != nil {
		logger.Error("failed-to-acquire-lock", err)
		retry.Reset(l.retryInterval)
	} else {
		logger.Info("acquired-lock")
		retry.Stop()
	}

	for {
		select {
		case sig := <-signals:
			logger.Info("signalled", lager.Data{"signal": sig})

			_, err := l.locker.Release(context.Background(), &models.ReleaseRequest{Resource: l.lock})
			if err != nil {
				logger.Error("failed-to-release-lock", err)
			} else {
				logger.Info("released-lock")
			}

			return nil

		case <-retry.C():
			_, err := l.locker.Lock(context.Background(), &models.LockRequest{Resource: l.lock})
			if err != nil {
				logger.Error("failed-to-acquire-lock", err)
				retry.Reset(l.retryInterval)
			} else {
				logger.Info("acquired-lock")
				retry.Stop()
			}
		}
	}

	return nil
}
