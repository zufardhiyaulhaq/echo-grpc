package settings

import "github.com/kelseyhightower/envconfig"

type Settings struct {
	Port string `envconfig:"GRPC_SERVER_PORT" default:"8081"`
}

func NewSettings() (Settings, error) {
	var settings Settings

	err := envconfig.Process("", &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
