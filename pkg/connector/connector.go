package connector

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"
)

type MattermostConnector struct {
	Bridge *bridgev2.Bridge
	Config Config
}

var _ bridgev2.NetworkConnector = (*MattermostConnector)(nil)

func (mc *MattermostConnector) Init(bridge *bridgev2.Bridge) {
	mc.Bridge = bridge
}

func (mc *MattermostConnector) Start(ctx context.Context) error {
	return nil
}

func (mc *MattermostConnector) GetName() bridgev2.BridgeName {
	return bridgev2.BridgeName{
		DisplayName:          "Mattermost",
		NetworkURL:           "https://mattermost.com/",
		NetworkID:            "mattermost",
		BeeperBridgeType:     "mattermost",
		DefaultPort:          29345,
		DefaultCommandPrefix: "!mm",
	}
}
