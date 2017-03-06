package locket

import (
	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientLocketConfig struct {
	LocketAddress    string `json:"locket_address,omitempty"`
	LocketCACert     string `json:"locket_ca_cert,omitempty"`
	LocketClientCert string `json:"locket_client_cert,omitempty"`
	LocketClientKey  string `json:"locket_client_key,omitempty"`
}

func NewClient(logger lager.Logger, config ClientLocketConfig) (models.LocketClient, error) {
	locketTLSConfig, err := cfhttp.NewTLSConfig(config.LocketClientCert, config.LocketClientKey, config.LocketCACert)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(config.LocketAddress, grpc.WithTransportCredentials(credentials.NewTLS(locketTLSConfig)))
	if err != nil {
		return nil, err
	}
	return models.NewLocketClient(conn), nil
}
