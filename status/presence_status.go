package status

import (
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type PresenceStatus struct {
	consul *consuladapter.Session
	logger lager.Logger
	clock  clock.Clock
}

func NewPresenceStatus(consul *consuladapter.Session, clock clock.Clock, logger lager.Logger) *PresenceStatus {
	return &PresenceStatus{
		consul: consul,
		logger: logger,
		clock:  clock,
	}
}
