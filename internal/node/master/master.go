package master

import (
	"github.com/actorgo-game/actorgo"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	app.Startup()
}
