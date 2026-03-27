package gate

import (
	"github.com/actorgo-game/actorgo"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)

	app.Register(newCheckCenter())
	app.Register(NewWSComponent())

	app.Startup()
}
