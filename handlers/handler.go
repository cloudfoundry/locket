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
	err := h.db.Lock(h.logger, req.Resource)
	if err != nil {
		return nil, err
	}
	return &models.LockResponse{}, nil
}

func (h *locketHandler) Release(ctx context.Context, req *models.ReleaseRequest) (*models.ReleaseResponse, error) {
	err := h.db.Release(h.logger, req.Resource)
	if err != nil {
		return nil, err
	}
	return &models.ReleaseResponse{}, nil
}

func (h *locketHandler) Fetch(ctx context.Context, req *models.FetchRequest) (*models.FetchResponse, error) {
	resource, err := h.db.Fetch(h.logger, req.Key)
	if err != nil {
		return nil, err
	}
	return &models.FetchResponse{
		Resource: resource,
	}, nil
}
