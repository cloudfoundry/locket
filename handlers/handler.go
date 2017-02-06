package handlers

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/models"
	"golang.org/x/net/context"
)

type locketHandler struct {
	db     db.LockDB
	logger lager.Logger
}

func NewLocketHandler(logger lager.Logger, db db.LockDB) *locketHandler {
	return &locketHandler{
		logger: logger,
		db:     db,
	}
}

func (h *locketHandler) Lock(ctx context.Context, req *models.LockRequest) (*models.LockResponse, error) {
	logger := h.logger.Session("lock", lager.Data{"req": req})
	logger.Info("started")
	defer logger.Info("complete")

	if req.TtlInSeconds <= 0 {
		return nil, models.ErrInvalidTTL
	}

	err := h.db.Lock(h.logger, req.Resource, req.TtlInSeconds)
	if err != nil {
		return nil, err
	}
	return &models.LockResponse{}, nil
}

func (h *locketHandler) Release(ctx context.Context, req *models.ReleaseRequest) (*models.ReleaseResponse, error) {
	logger := h.logger.Session("release", lager.Data{"request": req})
	logger.Info("started")
	defer logger.Info("complete")

	err := h.db.Release(h.logger, req.Resource)
	if err != nil {
		return nil, err
	}
	return &models.ReleaseResponse{}, nil
}

func (h *locketHandler) Fetch(ctx context.Context, req *models.FetchRequest) (*models.FetchResponse, error) {
	logger := h.logger.Session("fetch", lager.Data{"request": req})
	logger.Info("started")
	defer logger.Info("complete")

	resource, err := h.db.Fetch(h.logger, req.Key)
	if err != nil {
		return nil, err
	}
	return &models.FetchResponse{
		Resource: resource,
	}, nil
}
