package game

import (
	"github.com/actorgo-game/actorgo"
	"github.com/llr104/slgserver/internal/component"
	"github.com/llr104/slgserver/internal/node/game/module/mapmgr"
	"github.com/llr104/slgserver/internal/node/game/module/player"
	"github.com/llr104/slgserver/internal/node/game/module/unionmgr"
	"github.com/llr104/slgserver/internal/node/game/module/warmgr"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)

	app.Register(component.NewMongo())
	app.Register(component.NewRedis())

	app.AddActors(
		&player.ActorPlayers{},
		mapmgr.NewActorMap(),
		warmgr.NewActorWar(),
		unionmgr.NewActorUnion(),
	)

	app.Startup()
}
