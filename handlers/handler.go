package handlers

import (
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/expiration"
	metrics_helpers "code.cloudfoundry.org/locket/metrics/helpers"
	"code.cloudfoundry.org/locket/models"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

type locketHandler struct {
	logger lager.Logger

	db       db.LockDB
	exitCh   chan<- struct{}
	lockPick expiration.LockPick
	metrics  metrics_helpers.RequestMetrics
	models.UnimplementedLocketServer
}

func NewLocketHandler(logger lager.Logger, db db.LockDB, lockPick expiration.LockPick, requestMetrics metrics_helpers.RequestMetrics, exitCh chan<- struct{}) *locketHandler {
	return &locketHandler{
		logger:   logger,
		db:       db,
		lockPick: lockPick,
		exitCh:   exitCh,
		metrics:  requestMetrics,
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

func (h *locketHandler) monitorRequest(requestType string, ctx context.Context, key string, owner string, f func() error) error {
	h.metrics.IncrementRequestsStartedCounter(requestType, 1)
	h.metrics.IncrementRequestsInFlightCounter(requestType, 1)
	defer h.metrics.DecrementRequestsInFlightCounter(requestType, 1)

	start := time.Now()

	err := f()

	requestID := ""
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		requestID = md.Get("uuid")[0]
	}

	logData := lager.Data{
		"request-id":     requestID,
		"request-type":   requestType,
		"resource-key":   key,
		"resource-owner": owner,
	}
	if ctx.Err() == context.Canceled {
		h.logger.Info("context-cancelled", logData)
		h.metrics.IncrementRequestsCancelledCounter(requestType, 1)
	} else if ctx.Err() == context.DeadlineExceeded {
		h.logger.Info("context-deadline-exceeded", logData)
	}

	h.metrics.UpdateLatency(requestType, time.Since(start))

	if err != nil && err != models.ErrLockCollision {
		h.metrics.IncrementRequestsFailedCounter(requestType, 1)
		h.exitIfUnrecoverable(err)
	} else {
		h.metrics.IncrementRequestsSucceededCounter(requestType, 1)
	}
	return err
}

func (h *locketHandler) Lock(ctx context.Context, req *models.ProtoLockRequest) (*models.ProtoLockResponse, error) {
	var (
		response *models.ProtoLockResponse
		err      error
	)

	err = h.monitorRequest("Lock", ctx, req.Resource.Key, req.Resource.Owner, func() error {
		response, err = h.lock(ctx, req)
		return err
	})

	return response, err
}

func (h *locketHandler) Release(ctx context.Context, req *models.ProtoReleaseRequest) (*models.ProtoReleaseResponse, error) {
	var (
		response *models.ProtoReleaseResponse
		err      error
	)

	err = h.monitorRequest("Release", ctx, req.Resource.Key, req.Resource.Owner, func() error {
		response, err = h.release(ctx, req)
		return err
	})

	return response, err
}

func (h *locketHandler) Fetch(ctx context.Context, req *models.ProtoFetchRequest) (*models.ProtoFetchResponse, error) {
	var (
		response *models.ProtoFetchResponse
		err      error
	)

	err = h.monitorRequest("Fetch", ctx, req.Key, "", func() error {
		response, err = h.fetch(ctx, req)
		return err
	})

	return response, err
}

func (h *locketHandler) FetchAll(ctx context.Context, req *models.ProtoFetchAllRequest) (*models.ProtoFetchAllResponse, error) {
	var (
		response *models.ProtoFetchAllResponse
		err      error
	)

	err = h.monitorRequest("FetchAll", ctx, "", "", func() error {
		response, err = h.fetchAll(ctx, req)
		return err
	})

	return response, err
}

func (h *locketHandler) lock(ctx context.Context, req *models.ProtoLockRequest) (*models.ProtoLockResponse, error) {
	logger := h.logger.Session("lock")
	logger.Debug("started")
	defer logger.Debug("complete")

	err := validate(req)
	if err != nil {
		logger.Error("invalid-request", err, lager.Data{"typeCode": req.Resource.GetTypeCode()})

		return nil, err
	}

	if req.TtlInSeconds <= 0 {
		logger.Error("failed-locking-lock", models.ErrInvalidTTL, lager.Data{
			"key":   req.Resource.Key,
			"owner": req.Resource.Owner,
		})
		return nil, models.ErrInvalidTTL
	}

	if req.Resource.Owner == "" {
		logger.Error("failed-locking-lock", models.ErrInvalidOwner, lager.Data{
			"key":   req.Resource.Key,
			"owner": req.Resource.Owner,
		})
		return nil, models.ErrInvalidOwner
	}

	md, _ := metadata.FromIncomingContext(ctx)
	requestUUID := md["uuid"]
	if len(requestUUID) > 0 {
		logger = logger.WithData(lager.Data{"request-uuid": requestUUID[0]})
	}

	lock, err := h.db.Lock(ctx, logger, req.Resource.FromProto(), req.TtlInSeconds)
	if err != nil {
		if err != models.ErrLockCollision {
			logger.Error("failed-locking-lock", err, lager.Data{
				"key":   req.Resource.Key,
				"owner": req.Resource.Owner,
			})
		}
		return nil, err
	}

	h.lockPick.RegisterTTL(logger, lock)

	return &models.ProtoLockResponse{}, nil
}

func (h *locketHandler) release(ctx context.Context, req *models.ProtoReleaseRequest) (*models.ProtoReleaseResponse, error) {
	logger := h.logger.Session("release")
	logger.Debug("started")
	defer logger.Debug("complete")

	err := h.db.Release(ctx, logger, req.Resource.FromProto())
	if err != nil {
		return nil, err
	}

	return &models.ProtoReleaseResponse{}, nil
}

func (h *locketHandler) fetch(ctx context.Context, req *models.ProtoFetchRequest) (*models.ProtoFetchResponse, error) {
	logger := h.logger.Session("fetch")
	logger.Debug("started")
	defer logger.Debug("complete")

	lock, err := h.db.Fetch(ctx, logger, req.Key)
	if err != nil {
		return nil, err
	}

	return &models.ProtoFetchResponse{
		Resource: lock.Resource.ToProto(),
	}, nil
}

func (h *locketHandler) fetchAll(ctx context.Context, req *models.ProtoFetchAllRequest) (*models.ProtoFetchAllResponse, error) {
	logger := h.logger.Session("fetch-all")
	logger.Debug("started")
	defer logger.Debug("complete")

	err := validate(req)
	if err != nil {
		logger.Error("invalid-request", err, lager.Data{"typeCode": req.GetTypeCode()})
		return nil, err
	}

	resource := &models.ProtoResource{TypeCode: req.TypeCode}
	locks, err := h.db.FetchAll(ctx, logger, models.GetType(resource.FromProto()))
	if err != nil {
		return nil, err
	}

	var responses []*models.Resource
	for _, lock := range locks {
		responses = append(responses, lock.Resource)
	}

	response := &models.FetchAllResponse{
		Resources: responses,
	}
	return response.ToProto(), nil
}

func validate(req interface{}) error {
	var reqTypeCode models.TypeCode

	switch incomingReq := req.(type) {
	case *models.LockRequest:
		reqTypeCode = incomingReq.Resource.GetTypeCode()
	case *models.FetchAllRequest:
		reqTypeCode = incomingReq.GetTypeCode()
	default:
		return nil
	}

	if _, found := models.TypeCode_name[int32(reqTypeCode)]; !found {
		return models.ErrInvalidType
	}

	if reqTypeCode == models.UNKNOWN {
		return models.ErrInvalidType
	}

	return nil
}
