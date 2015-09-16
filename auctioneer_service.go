package locket

import "github.com/cloudfoundry-incubator/locket/presence"

func (l *client) AuctioneerAddress() (string, error) {
	value, err := l.consul.GetAcquiredValue(LockSchemaPath("auctioneer_lock"))
	if err != nil {
		return "", ErrServiceUnavailable
	}

	auctioneerPresence := presence.AuctioneerPresence{}
	err = presence.FromJSON(value, &auctioneerPresence)
	if err != nil {
		return "", err
	}

	return auctioneerPresence.AuctioneerAddress, nil
}
