package locket

import "github.com/cloudfoundry-incubator/runtime-schema/models"

func (l *Client) BBSMasterURL() (string, error) {
	value, err := l.consul.GetAcquiredValue(LockSchemaPath("bbs_lock"))
	if err != nil {
		return "", ErrServiceUnavailable
	}

	bbsPresence := models.BBSPresence{}
	err = models.FromJSON(value, &bbsPresence)
	if err != nil {
		return "", err
	}

	return bbsPresence.URL, nil
}
