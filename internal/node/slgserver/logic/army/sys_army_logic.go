package army

import (
	"sync"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/global"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/npc"
)

func NewSysArmy() *sysArmyLogic {
	return &sysArmyLogic{
		sysArmys: make(map[int][]*model.Army),
	}
}

type sysArmyLogic struct {
	mutex    sync.Mutex
	sysArmys map[int][]*model.Army // key: posId 系统建筑军队
}

func (this *sysArmyLogic) GetArmy(x, y int) []*model.Army {
	posId := global.ToPosition(x, y)

	this.mutex.Lock()
	a, ok := this.sysArmys[posId]
	this.mutex.Unlock()
	if ok {
		return a
	}

	armys := make([]*model.Army, 0)
	mapBuild, ok := mgr.NMMgr.PositionBuild(x, y)
	if !ok {
		return armys
	}
	cfg, ok := static_conf.MapBuildConf.BuildConfig(mapBuild.Type, mapBuild.Level)
	if !ok {
		return armys
	}
	soldiers := npc.Cfg.NPCSoilder(cfg.Level)
	hasArmy, armyCfg := npc.Cfg.RandomOne(cfg.Level)
	if !hasArmy {
		return armys
	}
	out, ok := mgr.GMgr.GetNPCGenerals(armyCfg.CfgIds, armyCfg.Lvs)
	if !ok {
		return armys
	}

	gsId := [model.ArmyGCnt]int{}
	gs := [model.ArmyGCnt]*model.General{}
	for i := 0; i < len(out) && i < model.ArmyGCnt; i++ {
		gs[i] = &out[i]
		gsId[i] = out[i].Id
	}
	scnt := [model.ArmyGCnt]int{soldiers, soldiers, soldiers}

	a2 := &model.Army{
		RId: 0, Order: 0, CityId: 0,
		GeneralArray: gsId, Gens: gs, SoldierArray: scnt,
	}
	armys = append(armys, a2)

	this.mutex.Lock()
	this.sysArmys[posId] = armys
	this.mutex.Unlock()
	return armys
}

func (this *sysArmyLogic) DelArmy(x, y int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	posId := global.ToPosition(x, y)
	delete(this.sysArmys, posId)
}
