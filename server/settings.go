package main

import "github.com/kelseyhightower/envconfig"

type Settings struct {
	Port int `envconfig:"PORT" default:"8080"`
}

func NewSettings() (Settings, error) {
	var settings Settings

	err := envconfig.Process("", &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
