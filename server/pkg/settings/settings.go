package settings

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	Port                 string        `envconfig:"GRPC_SERVER_PORT" default:"8081"`
	GRPCKeepalive        bool          `envconfig:"GRPC_SERVER_KEEPALIVE" default:"false"`
	GRPCKeepaliveTime    time.Duration `envconfig:"GRPC_SERVER_KEEPALIVE_TIME" default:"10s"`
	GRPCKeepaliveTimeout time.Duration `envconfig:"GRPC_SERVER_KEEPALIVE_TIMEOUT" default:"20s"`
}

func NewSettings() (Settings, error) {
	var settings Settings

	err := envconfig.Process("", &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
