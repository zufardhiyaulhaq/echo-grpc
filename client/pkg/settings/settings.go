package settings

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	EchoPort             string        `envconfig:"ECHO_PORT" default:"5000"`
	HTTPPort             string        `envconfig:"HTTP_PORT" default:"80"`
	GRPCKeepalive        bool          `envconfig:"GRPC_CLIENT_KEEPALIVE" default:"false"`
	GRPCKeepaliveTime    time.Duration `envconfig:"GRPC_CLIENT_KEEPALIVE_TIME" default:"10s"`
	GRPCKeepaliveTimeout time.Duration `envconfig:"GRPC_CLIENT_KEEPALIVE_TIMEOUT" default:"20s"`
	GRPCServerHost       string        `envconfig:"GRPC_SERVER_HOST" default:"server"`
	GRPCServerPort       string        `envconfig:"GRPC_SERVER_PORT" default:"8080"`
	GRPCServerTLS        bool          `envconfig:"GRPC_SERVER_TLS" default:false`
}

func NewSettings() (Settings, error) {
	var settings Settings

	err := envconfig.Process("", &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
