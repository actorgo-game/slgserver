package slgserver

import (
	"github.com/actorgo-game/actorgo"
	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	credis "github.com/actorgo-game/actorgo/components/redis"
	cfacade "github.com/actorgo-game/actorgo/facade"
	cserializer "github.com/actorgo-game/actorgo/net/serializer"

	"github.com/llr104/slgserver/internal/node/slgserver/controller"
	"github.com/llr104/slgserver/internal/node/slgserver/run"
)

const slgInitComponentName = "slgserver_init"

type slgInitComponent struct {
	cfacade.Component
	dbID     string
	serverId int
}

func (p *slgInitComponent) Name() string {
	return slgInitComponentName
}

func (p *slgInitComponent) OnAfterInit() {
	run.Init(p.App(), p.dbID, p.serverId)
}

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	app.SetSerializer(cserializer.NewJSON())

	app.Register(cmongo.NewComponent())
	app.Register(credis.NewComponent())
	app.Register(&slgInitComponent{dbID: "slg_db", serverId: 0})

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
