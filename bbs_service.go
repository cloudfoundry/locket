package locket

import "github.com/cloudfoundry-incubator/locket/presence"

func (l *client) BBSMasterURL() (string, error) {
	value, err := l.consul.GetAcquiredValue(LockSchemaPath("bbs_lock"))
	if err != nil {
		return "", ErrServiceUnavailable
	}

	bbsPresence := presence.BBSPresence{}
	err = presence.FromJSON(value, &bbsPresence)
	if err != nil {
		return "", err
	}

	return bbsPresence.URL, nil
}
