package loginserver

import (
	"github.com/actorgo-game/actorgo"
	"github.com/llr104/slgserver/internal/component"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	app.Register(component.NewMongo())
	app.Register(component.NewRedis())
	app.AddActors(
		&ActorAccount{},
		&ActorOps{},
	)
	app.Startup()
}
