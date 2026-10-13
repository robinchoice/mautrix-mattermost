package connector

import (
	_ "embed"

	up "go.mau.fi/util/configupgrade"
)

//go:embed example-config.yaml
var ExampleConfig string

type Config struct{}

func (mc *MattermostConnector) GetConfig() (example string, data any, upgrader up.Upgrader) {
	return ExampleConfig, &mc.Config, &up.StructUpgrader{
		SimpleUpgrader: func(helper up.Helper) {},
		Base:           ExampleConfig,
	}
}
