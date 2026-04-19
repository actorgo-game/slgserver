package run

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"

	"github.com/llr104/slgserver/internal/node/slgserver/db"
	"github.com/llr104/slgserver/internal/node/slgserver/logic"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/general"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/npc"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/skill"
)

// Init 在 slgserver 节点 actorgo 完成 Configure 之后、Startup 之前调用：
//
//  1. 加载所有静态 JSON 配置；
//  2. 初始化 mongo 数据库句柄；
//  3. 通过 mgr.Init() 注册 model 钩子并装载所有数据；
//  4. 通过 logic.BeforeInit/Init/AfterInit 启动业务逻辑层。
func Init(app cfacade.IApplication, dbId string, serverId int) {
	loadStaticConf()
	db.Setup(dbId, serverId)
	mgr.Init()
	logic.BeforeInit()
	logic.Init()
	logic.AfterInit()
	clog.Info("[slgserver/run] init finished: db=%s, serverId=%d", dbId, serverId)
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
