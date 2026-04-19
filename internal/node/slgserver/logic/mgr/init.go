package mgr

import (
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/persistence"
)

// Init 由 slgserver 节点在 db.Setup 之后、controller 注册之前调用：
//
//  1. 启动异步落库 worker；
//  2. 把 SyncWriter 钩子指向 worker；
//  3. 注入 model 层依赖的配置/查询钩子；
//  4. 按依赖顺序加载所有 mgr 缓存。
//
// 注意：UnionMgr 必须在 RAttrMgr 之前 Load，因为 RAttrMgr.Load 需要回填 UnionId。
func Init() {
	persistence.Setup()
	model.SyncWriter = persistence.Submit

	registerHooks()

	// 配置/角色相关
	NMMgr.Load()
	UnionMgr.Load()
	RMgr.Load()
	RAttrMgr.Load()
	RResMgr.Load()
	RFMgr.Load()
	RCMgr.Load()
	RBMgr.Load()
	GMgr.Load()
	SkillMgr.Load()
	AMgr.Load()
}

// registerHooks 把 model 层 var 函数指针指向 mgr 内部实现，等价于
// 原 slgserver/logic.BeforeInit。
func registerHooks() {
	model.GetRoleNickName = RoleNickName
	model.GetUnionId = UnionId
	model.GetUnionName = UnionName
	model.GetParentId = ParentId
	model.GetMainMembers = MainMembers
	model.GetYield = GetYield
	model.GetDepotCapacity = GetDepotCapacity
	model.GetCityCost = GetCityCost
	model.GetMaxDurable = GetMaxDurable
	model.GetCityLv = GetCityLV
	model.MapResTypeLevel = NMMgr.MapResTypeLevel
	model.FacilityCostTime = FacilityCostTime
	model.GetSkillCfg = SkillCfgInjector
	model.GetGeneralCfg = GeneralCfgInjector
	model.NewGeneralBaseInit = NewGeneralBaseInit

	// CoalitionLog Insert hook：直接走 persistence
	model.CoalitionLogInsertHook = func(l *model.CoalitionLog) {
		persistence.Submit(l)
	}
}
