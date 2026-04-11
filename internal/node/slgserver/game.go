package slgserver

import (
	"github.com/actorgo-game/actorgo"
	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	credis "github.com/actorgo-game/actorgo/components/redis"
	"github.com/llr104/slgserver/internal/node/slgserver/module/mapmgr"
	"github.com/llr104/slgserver/internal/node/slgserver/module/player"
	"github.com/llr104/slgserver/internal/node/slgserver/module/unionmgr"
	"github.com/llr104/slgserver/internal/node/slgserver/module/warmgr"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)

	app.Register(cmongo.NewComponent())
	app.Register(credis.NewComponent())

	app.AddActors(
		&player.ActorPlayers{},
		mapmgr.NewActorMap(),
		warmgr.NewActorWar(),
		unionmgr.NewActorUnion(),
	)

	app.Startup()
}
