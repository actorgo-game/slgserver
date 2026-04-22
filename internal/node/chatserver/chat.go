package chatserver

import (
	"github.com/actorgo-game/actorgo"
	cserializer "github.com/actorgo-game/actorgo/net/serializer"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	app.SetSerializer(cserializer.NewJSON())

	app.AddActors(
		NewActorChatRoom(),
	)

	app.Startup()
}
