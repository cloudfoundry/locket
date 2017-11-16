package metrics

import "code.cloudfoundry.org/lager"

//go:generate counterfeiter . LockDBMetrics
type LockDBMetrics interface {
	OpenConnections() int
	Count(logger lager.Logger, lockType string) (int, error)
}
