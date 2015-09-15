package locket

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

func (l *Locket) NewCellPresence(cellPresence models.CellPresence, retryInterval time.Duration) ifrit.Runner {
	payload, err := models.ToJSON(cellPresence)
	if err != nil {
		panic(err)
	}

	return maintainer.NewPresence(l.consul, shared.CellSchemaPath(cellPresence.CellID), payload, l.clock, retryInterval, l.logger)
}

func (l *Locket) CellById(cellId string) (models.CellPresence, error) {
	cellPresence := models.CellPresence{}

	value, err := l.consul.GetAcquiredValue(shared.CellSchemaPath(cellId))
	if err != nil {
		return cellPresence, shared.ConvertConsulError(err)
	}

	err = models.FromJSON(value, &cellPresence)
	if err != nil {
		return cellPresence, err
	}

	return cellPresence, nil
}

func (l *Locket) Cells() ([]models.CellPresence, error) {
	cells, err := l.consul.ListAcquiredValues(shared.CellSchemaRoot)
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
			l.logger.Error("failed-to-unmarshal-cells-json", err)
			continue
		}

		cellPresences = append(cellPresences, cellPresence)
	}

	return cellPresences, nil
}

func (l *Locket) CellEvents() <-chan CellEvent {
	logger := l.logger.Session("cell-events")

	events := make(chan CellEvent)
	go func() {
		disappeared := l.consul.WatchForDisappearancesUnder(logger, shared.CellSchemaRoot)

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
