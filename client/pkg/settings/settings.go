package settings

import "github.com/kelseyhightower/envconfig"

type Settings struct {
	EchoPort       string `envconfig:"ECHO_PORT" default:"5000"`
	HTTPPort       string `envconfig:"HTTP_PORT" default:"80"`
	GRPCServerHost string `envconfig:"GRPC_SERVER_HOST" default:"server"`
	GRPCServerPort string `envconfig:"GRPC_SERVER_PORT" default:"8080"`
	GRPCServerTLS  bool   `envconfig:"GRPC_SERVER_TLS" default:false`
}

func NewSettings() (Settings, error) {
	var settings Settings

	err := envconfig.Process("", &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
