package status

import (
	"github.com/cloudfoundry-incubator/locket/shared"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func (p *PresenceStatus) AuctioneerAddress() (string, error) {
	value, err := p.consul.GetAcquiredValue(shared.LockSchemaPath("auctioneer_lock"))
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
