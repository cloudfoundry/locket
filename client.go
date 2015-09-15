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

type Client struct {
	consul *consuladapter.Session
	logger lager.Logger
	clock  clock.Clock
}

func NewClient(consul *consuladapter.Session, clock clock.Clock, logger lager.Logger) *Client {
	return &Client{
		consul: consul,
		logger: logger,
		clock:  clock,
	}
}

func (l *Client) NewAuctioneerLock(auctioneerPresence models.AuctioneerPresence, retryInterval time.Duration) (ifrit.Runner, error) {
	auctionerPresenceJSON, err := models.ToJSON(auctioneerPresence)
	if err != nil {
		return nil, err
	}
	return maintainer.NewLock(l.consul, LockSchemaPath("auctioneer_lock"), auctionerPresenceJSON, l.clock, retryInterval, l.logger), nil
}

func (l *Client) NewConvergeLock(convergerID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("converge_lock"), []byte(convergerID), l.clock, retryInterval, l.logger)
}

func (l *Client) NewNsyncBulkerLock(bulkerID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("nsync_bulker_lock"), []byte(bulkerID), l.clock, retryInterval, l.logger)
}

func (l *Client) NewRouteEmitterLock(emitterID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("route_emitter_lock"), []byte(emitterID), l.clock, retryInterval, l.logger)
}

func (l *Client) NewRuntimeMetricsLock(runtimeMetricsID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("runtime_metrics_lock"), []byte(runtimeMetricsID), l.clock, retryInterval, l.logger)
}

func (l *Client) NewTpsWatcherLock(tpsWatcherID string, retryInterval time.Duration) ifrit.Runner {
	return maintainer.NewLock(l.consul, LockSchemaPath("tps_watcher_lock"), []byte(tpsWatcherID), l.clock, retryInterval, l.logger)
}

func (l *Client) NewBBSMasterLock(bbsPresence models.BBSPresence, retryInterval time.Duration) (ifrit.Runner, error) {
	bbsPresenceJSON, err := models.ToJSON(bbsPresence)
	if err != nil {
		return nil, err
	}
	return maintainer.NewLock(l.consul, LockSchemaPath("bbs_lock"), bbsPresenceJSON, l.clock, retryInterval, l.logger), nil
}
