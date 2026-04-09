package chatserver

import (
	"github.com/actorgo-game/actorgo"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)

	app.AddActors(
		NewActorChatRoom(),
	)

	app.Startup()
}
