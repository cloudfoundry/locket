package locket

import (
	"path"
	"time"

	"github.com/cloudfoundry-incubator/locket/maintainer"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

const CellPresenceTTL = 10 * time.Second

func (l *client) NewCellPresence(cellPresence models.CellPresence, retryInterval time.Duration) ifrit.Runner {
	payload, err := models.ToJSON(cellPresence)
	if err != nil {
		panic(err)
	}

	return maintainer.NewPresence(l.consul, CellSchemaPath(cellPresence.CellID), payload, l.clock, retryInterval, l.logger)
}

func (l *client) CellById(cellId string) (models.CellPresence, error) {
	cellPresence := models.CellPresence{}

	value, err := l.consul.GetAcquiredValue(CellSchemaPath(cellId))
	if err != nil {
		return cellPresence, ConvertConsulError(err)
	}

	err = models.FromJSON(value, &cellPresence)
	if err != nil {
		return cellPresence, err
	}

	return cellPresence, nil
}

func (l *client) Cells() ([]models.CellPresence, error) {
	cells, err := l.consul.ListAcquiredValues(CellSchemaRoot)
	if err != nil {
		err = ConvertConsulError(err)
		if err != ErrStoreResourceNotFound {
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

func (l *client) CellEvents() <-chan CellEvent {
	logger := l.logger.Session("cell-events")

	events := make(chan CellEvent)
	go func() {
		disappeared := l.consul.WatchForDisappearancesUnder(logger, CellSchemaRoot)

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
