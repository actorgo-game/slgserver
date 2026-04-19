package slgserver

import (
	"github.com/actorgo-game/actorgo"
	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	credis "github.com/actorgo-game/actorgo/components/redis"
	"github.com/llr104/slgserver/internal/node/slgserver/module/mapmgr"
	"github.com/llr104/slgserver/internal/node/slgserver/module/player"
	"github.com/llr104/slgserver/internal/node/slgserver/module/unionmgr"
	"github.com/llr104/slgserver/internal/node/slgserver/module/warmgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/general"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/npc"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/skill"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)

	app.Register(cmongo.NewComponent())
	app.Register(credis.NewComponent())

	loadStaticConf()

	app.AddActors(
		&player.ActorPlayers{},
		mapmgr.NewActorMap(),
		warmgr.NewActorWar(),
		unionmgr.NewActorUnion(),
	)

	app.Startup()
}

func loadStaticConf() {
	static_conf.Basic.Load()
	static_conf.MapBuildConf.Load()
	static_conf.MapBCConf.Load()

	facility.FConf.Load()
	general.GenBasic.Load()
	skill.Skill.Load()
	general.General.Load()
	npc.Cfg.Load()

}
