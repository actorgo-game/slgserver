package slgserver

import (
	"github.com/actorgo-game/actorgo"
	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	credis "github.com/actorgo-game/actorgo/components/redis"

	"github.com/llr104/slgserver/internal/node/slgserver/controller"
	"github.com/llr104/slgserver/internal/node/slgserver/run"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)

	app.Register(cmongo.NewComponent())
	app.Register(credis.NewComponent())

	run.Init(app, "slg_db", 0)

	app.AddActors(
		controller.NewActorRole(),
		controller.NewActorCity(),
		controller.NewActorNationMap(),
		controller.NewActorGeneral(),
		controller.NewActorSkill(),
		controller.NewActorArmy(),
		controller.NewActorWar(),
		controller.NewActorUnion(),
		controller.NewActorInterior(),
	)

	app.Startup()
}
