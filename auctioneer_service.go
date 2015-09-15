package locket

import (
	"github.com/cloudfoundry-incubator/locket/shared"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func (l *Locket) AuctioneerAddress() (string, error) {
	value, err := l.consul.GetAcquiredValue(shared.LockSchemaPath("auctioneer_lock"))
	if err != nil {
		return "", shared.ErrServiceUnavailable
	}

	auctioneerPresence := models.AuctioneerPresence{}
	err = models.FromJSON(value, &auctioneerPresence)
	if err != nil {
		return "", err
	}

	return auctioneerPresence.AuctioneerAddress, nil
}
