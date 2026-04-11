package loginserver

import (
	"github.com/actorgo-game/actorgo"
	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	credis "github.com/actorgo-game/actorgo/components/redis"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	app.Register(cmongo.NewComponent())
	app.Register(credis.NewComponent())
	app.AddActors(
		&ActorAccount{},
		&ActorOps{},
	)
	app.Startup()
}
