package main

import (
	"maunium.net/go/mautrix/bridgev2/matrix/mxmain"

	"github.com/robinchoice/mautrix-mattermost/pkg/connector"
)

// Filled at build time with -X linker flags.
var (
	Tag       = "unknown"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	m := mxmain.BridgeMain{
		Name:        "mautrix-mattermost",
		Description: "A Matrix-Mattermost puppeting bridge",
		URL:         "https://github.com/robinchoice/mautrix-mattermost",
		Version:     "0.1.0",
		Connector:   &connector.MattermostConnector{},
	}
	m.InitVersion(Tag, Commit, BuildTime)
	m.Run()
}
