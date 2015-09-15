package locket

import "github.com/cloudfoundry-incubator/runtime-schema/models"

func (l *client) AuctioneerAddress() (string, error) {
	value, err := l.consul.GetAcquiredValue(LockSchemaPath("auctioneer_lock"))
	if err != nil {
		return "", ErrServiceUnavailable
	}

	auctioneerPresence := models.AuctioneerPresence{}
	err = models.FromJSON(value, &auctioneerPresence)
	if err != nil {
		return "", err
	}

	return auctioneerPresence.AuctioneerAddress, nil
}
