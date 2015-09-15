package locket

import (
	"time"

	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/locket/maintainer"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

//go:generate counterfeiter -o locketfakes/fake_client.go . Client
type Client interface {
	// Locks
	NewAuctioneerLock(auctioneerPresence models.AuctioneerPresence, retryInterval time.Duration) (ifrit.Runner, error)
	NewConvergeLock(convergerID string, retryInterval time.Duration) ifrit.Runner
	NewNsyncBulkerLock(bulkerID string, retryInterval time.Duration) ifrit.Runner
	NewRouteEmitterLock(emitterID string, retryInterval time.Duration) ifrit.Runner
	NewRuntimeMetricsLock(runtimeMetricsID string, retryInterval time.Duration) ifrit.Runner
	NewTpsWatcherLock(tpsWatcherID string, retryInterval time.Duration) ifrit.Runner
	NewBBSMasterLock(bbsPresence models.BBSPresence, retryInterval time.Duration) (ifrit.Runner, error)

	// Presence
	AuctioneerAddress() (string, error)
	BBSMasterURL() (string, error)

	NewCellPresence(cellPresence models.CellPresence, retryInterval time.Duration) ifrit.Runner
	CellById(cellId string) (models.CellPresence, error)
	Cells() ([]models.CellPresence, error)
	CellEvents() <-chan CellEvent
}

type client struct {
	consul *consuladapter.Session
	logger lager.Logger
	clock  clock.Clock
}

func NewClient(consul *consuladapter.Session, clock clock.Clock, logger lager.Logger) Client {
	return &client{
		consul: consul,
		logger: logger,
		clock:  clock,
	}
}

func (l *client) NewAuctioneerLock(auctioneerPresence models.AuctioneerPresence, retryInterval time.Duration) (ifrit.Runner, error) {
	auctionerPresenceJSON, err := models.ToJSON(auctioneerPresence)
	if err != nil {
		return nil, err
	}
	return maintainer.NewLock(l.consul, LockSchemaPath("auctioneer_lock"), auctionerPresenceJSON, l.clock, retryInterval, l.logger), nil
}

func (l *client) NewConvergeLock(convergerID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("converge_lock"), []byte(convergerID), l.clock, retryInterval, l.logger)
}

func (l *client) NewNsyncBulkerLock(bulkerID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("nsync_bulker_lock"), []byte(bulkerID), l.clock, retryInterval, l.logger)
}

func (l *client) NewRouteEmitterLock(emitterID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("route_emitter_lock"), []byte(emitterID), l.clock, retryInterval, l.logger)
}

func (l *client) NewRuntimeMetricsLock(runtimeMetricsID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("runtime_metrics_lock"), []byte(runtimeMetricsID), l.clock, retryInterval, l.logger)
}

func (l *client) NewTpsWatcherLock(tpsWatcherID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("tps_watcher_lock"), []byte(tpsWatcherID), l.clock, retryInterval, l.logger)
}

func (l *client) NewBBSMasterLock(bbsPresence models.BBSPresence, retryInterval time.Duration) (ifrit.Runner, error) {
	bbsPresenceJSON, err := models.ToJSON(bbsPresence)
	if err != nil {
		return nil, err
	}
	return maintainer.NewLock(l.consul, LockSchemaPath("bbs_lock"), bbsPresenceJSON, l.clock, retryInterval, l.logger), nil
}
