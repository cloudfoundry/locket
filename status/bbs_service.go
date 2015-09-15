package status

import (
	"github.com/cloudfoundry-incubator/locket/shared"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func (p *PresenceStatus) BBSMasterURL() (string, error) {
	value, err := p.consul.GetAcquiredValue(shared.LockSchemaPath("bbs_lock"))
	if err != nil {
		return "", shared.ErrServiceUnavailable
	}

	bbsPresence := models.BBSPresence{}
	err = models.FromJSON(value, &bbsPresence)
	if err != nil {
		return "", err
	}

	return bbsPresence.URL, nil
}
