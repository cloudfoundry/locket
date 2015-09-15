package status

import (
	"path"
	"time"

	"github.com/cloudfoundry-incubator/locket/maintainer"
	"github.com/cloudfoundry-incubator/locket/shared"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

const CellPresenceTTL = 10 * time.Second

func (p *PresenceStatus) NewCellPresence(cellPresence models.CellPresence, retryInterval time.Duration) ifrit.Runner {
	payload, err := models.ToJSON(cellPresence)
	if err != nil {
		panic(err)
	}

	return maintainer.NewPresence(p.consul, shared.CellSchemaPath(cellPresence.CellID), payload, p.clock, retryInterval, p.logger)
}

func (p *PresenceStatus) CellById(cellId string) (models.CellPresence, error) {
	cellPresence := models.CellPresence{}

	value, err := p.consul.GetAcquiredValue(shared.CellSchemaPath(cellId))
	if err != nil {
		return cellPresence, shared.ConvertConsulError(err)
	}

	err = models.FromJSON(value, &cellPresence)
	if err != nil {
		return cellPresence, err
	}

	return cellPresence, nil
}

func (p *PresenceStatus) Cells() ([]models.CellPresence, error) {
	cells, err := p.consul.ListAcquiredValues(shared.CellSchemaRoot)
	if err != nil {
		err = shared.ConvertConsulError(err)
		if err != shared.ErrStoreResourceNotFound {
			return nil, err
		}
	}

	cellPresences := []models.CellPresence{}
	for _, cell := range cells {
		cellPresence := models.CellPresence{}
		err := models.FromJSON(cell, &cellPresence)
		if err != nil {
			p.logger.Error("failed-to-unmarshal-cells-json", err)
			continue
		}

		cellPresences = append(cellPresences, cellPresence)
	}

	return cellPresences, nil
}

func (p *PresenceStatus) CellEvents() <-chan CellEvent {
	logger := p.logger.Session("cell-events")

	events := make(chan CellEvent)
	go func() {
		disappeared := p.consul.WatchForDisappearancesUnder(logger, shared.CellSchemaRoot)

		for {
			select {
			case keys, ok := <-disappeared:
				if !ok {
					return
				}

				cellIDs := make([]string, len(keys))
				for i, key := range keys {
					cellIDs[i] = path.Base(key)
				}
				e := CellDisappearedEvent{cellIDs}

				logger.Info("cell-disappeared", lager.Data{"cell-ids": e.CellIDs()})
				events <- e
			}
		}
	}()

	return events
}

type CellEvent interface {
	EventType() CellEventType
	CellIDs() []string
}

type CellEventType int

const (
	CellEventTypeInvalid CellEventType = iota
	CellDisappeared
)

type CellDisappearedEvent struct {
	IDs []string
}

func (CellDisappearedEvent) EventType() CellEventType {
	return CellDisappeared
}

func (e CellDisappearedEvent) CellIDs() []string {
	return e.IDs
}
