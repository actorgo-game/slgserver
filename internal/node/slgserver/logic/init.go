package logic

import (
	"time"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/army"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/union"
)

var (
	Union     *union.UnionLogic
	ArmyLogic *army.ArmyLogic
)

// BeforeInit 注入需要 logic 包提供的钩子。绝大部分钩子已在 mgr.Init() 注册，
// 这里只补充依赖 logic/army 包的 ArmyIsInView。
func BeforeInit() {
	model.ArmyIsInView = army.ArmyIsInView
}

// Init 初始化各 logic 单例。
func Init() {
	Union = union.Instance()
	ArmyLogic = army.Instance()
}

// AfterInit 启动 logic 层的后台任务（建筑放弃 / 摧毁的轮询）。
func AfterInit() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			buildIds := mgr.RBMgr.CheckGiveUp()
			for _, posId := range buildIds {
				ArmyLogic.GiveUp(posId)
			}

			buildIds = mgr.RBMgr.CheckDestroy()
			for _, posId := range buildIds {
				ArmyLogic.Interrupt(posId)
			}
		}
	}()
}
