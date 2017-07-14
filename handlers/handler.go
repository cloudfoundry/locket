package handlers

import (
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/expiration"
	"code.cloudfoundry.org/locket/models"
	"golang.org/x/net/context"
)

type locketHandler struct {
	logger lager.Logger

	db       db.LockDB
	exitCh   chan<- struct{}
	lockPick expiration.LockPick
}

func NewLocketHandler(logger lager.Logger, db db.LockDB, lockPick expiration.LockPick, exitCh chan<- struct{}) *locketHandler {
	return &locketHandler{
		logger:   logger,
		db:       db,
		lockPick: lockPick,
		exitCh:   exitCh,
	}
}

func (h *locketHandler) exitIfUnrecoverable(err error) {
	if err != helpers.ErrUnrecoverableError {
		return
	}

	h.logger.Error("unrecoverable-error", err)

	select {
	case h.exitCh <- struct{}{}:
	default:
	}
}

func (h *locketHandler) Lock(ctx context.Context, req *models.LockRequest) (*models.LockResponse, error) {
	logger := h.logger.Session("lock")
	logger.Debug("started")
	defer logger.Debug("complete")

	if req.TtlInSeconds <= 0 {
		return nil, models.ErrInvalidTTL
	}

	if req.Resource.Owner == "" {
		return nil, models.ErrInvalidOwner
	}

	lock, err := h.db.Lock(logger, req.Resource, req.TtlInSeconds)
	if err != nil {
		h.exitIfUnrecoverable(err)
		if err != models.ErrLockCollision {
			h.logger.Error("failed-locking-lock", err)
		}
		return nil, err
	}

	h.lockPick.RegisterTTL(logger, lock)

	return &models.LockResponse{}, nil
}

func (h *locketHandler) Release(ctx context.Context, req *models.ReleaseRequest) (*models.ReleaseResponse, error) {
	logger := h.logger.Session("release")
	logger.Debug("started")
	defer logger.Debug("complete")

	err := h.db.Release(logger, req.Resource)
	if err != nil {
		h.exitIfUnrecoverable(err)
		return nil, err
	}
	return &models.ReleaseResponse{}, nil
}

func (h *locketHandler) Fetch(ctx context.Context, req *models.FetchRequest) (*models.FetchResponse, error) {
	logger := h.logger.Session("fetch")
	logger.Debug("started")
	defer logger.Debug("complete")

	lock, err := h.db.Fetch(logger, req.Key)
	if err != nil {
		h.exitIfUnrecoverable(err)
		return nil, err
	}
	return &models.FetchResponse{
		Resource: lock.Resource,
	}, nil
}

func (h *locketHandler) FetchAll(ctx context.Context, req *models.FetchAllRequest) (*models.FetchAllResponse, error) {
	logger := h.logger.Session("fetch-all")
	logger.Debug("started")
	defer logger.Debug("complete")

	locks, err := h.db.FetchAll(logger, req.Type)
	if err != nil {
		h.exitIfUnrecoverable(err)
		return nil, err
	}

	var responses []*models.Resource
	for _, lock := range locks {
		responses = append(responses, lock.Resource)
	}

	return &models.FetchAllResponse{
		Resources: responses,
	}, nil
}
