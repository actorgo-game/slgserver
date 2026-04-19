package mgr

import (
	"sync"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/general"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/npc"
	"github.com/llr104/slgserver/internal/util"
)

type generalMgr struct {
	mutex     sync.RWMutex
	genByRole map[int][]*model.General // key: rid
	genByGId  map[int]*model.General   // key: id
}

var GMgr = &generalMgr{
	genByRole: make(map[int][]*model.General),
	genByGId:  make(map[int]*model.General),
}

func (this *generalMgr) updatePhysicalPower() {
	limit := static_conf.Basic.General.PhysicalPowerLimit
	recoverCnt := static_conf.Basic.General.RecoveryPhysicalPower
	for {
		time.Sleep(1 * time.Hour)
		this.mutex.RLock()
		for _, g := range this.genByGId {
			if g.PhysicalPower < limit {
				g.PhysicalPower = util.MinInt(limit, g.PhysicalPower+recoverCnt)
				g.SyncExecute()
			}
		}
		this.mutex.RUnlock()
	}
}

// createNPC 创建 NPC 武将（rid=0）。
func (this *generalMgr) createNPC() ([]*model.General, bool) {
	gs := make([]*model.General, 0)
	for _, armys := range npc.Cfg.Armys {
		for _, cfgs := range armys.ArmyCfg {
			for i, cfgId := range cfgs.CfgIds {
				r, ok := this.NewGeneral(cfgId, 0, cfgs.Lvs[i])
				if !ok {
					return nil, false
				}
				gs = append(gs, r)
			}
		}
	}
	return gs, true
}

func (this *generalMgr) add(g *model.General) {
	this.mutex.Lock()
	if _, ok := this.genByRole[g.RId]; !ok {
		this.genByRole[g.RId] = make([]*model.General, 0)
	}
	this.genByRole[g.RId] = append(this.genByRole[g.RId], g)
	this.genByGId[g.Id] = g
	this.mutex.Unlock()
}

func (this *generalMgr) Load() {
	c := coll(model.General{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()

	cursor, err := c.Find(cx, withServer(bson.M{"state": model.GeneralNormal}))
	if err != nil {
		clog.Warn("[GMgr] Load find err=%v", err)
		return
	}
	var rows []*model.General
	if err := cursor.All(cx, &rows); err != nil {
		clog.Warn("[GMgr] Load decode err=%v", err)
		return
	}
	for _, v := range rows {
		this.genByGId[v.Id] = v
		if _, ok := this.genByRole[v.RId]; !ok {
			this.genByRole[v.RId] = make([]*model.General, 0)
		}
		this.genByRole[v.RId] = append(this.genByRole[v.RId], v)
	}

	if len(this.genByGId) == 0 {
		this.createNPC()
	}

	go this.updatePhysicalPower()
}

func (this *generalMgr) GetByRId(rid int) ([]*model.General, bool) {
	this.mutex.Lock()
	r, ok := this.genByRole[rid]
	this.mutex.Unlock()
	if ok {
		out := make([]*model.General, 0, len(r))
		for _, g := range r {
			if g.IsActive() {
				out = append(out, g)
			}
		}
		return out, true
	}

	c := coll(model.General{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{"rid": rid, "state": model.GeneralNormal}))
	if err != nil {
		clog.Warn("[GMgr] GetByRId find err=%v rid=%d", err, rid)
		return nil, false
	}
	var gs []*model.General
	if err := cursor.All(cx, &gs); err != nil {
		return nil, false
	}
	if len(gs) == 0 {
		return nil, false
	}
	for _, g := range gs {
		this.add(g)
	}
	return gs, true
}

func (this *generalMgr) GetByGId(gid int) (*model.General, bool) {
	this.mutex.RLock()
	g, ok := this.genByGId[gid]
	this.mutex.RUnlock()
	if ok {
		if g.IsActive() {
			return g, true
		}
		return nil, false
	}

	c := coll(model.General{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	g = &model.General{}
	if err := c.FindOne(cx, withServer(bson.M{"id": gid, "state": model.GeneralNormal})).Decode(g); err != nil {
		return nil, false
	}
	this.add(g)
	return g, true
}

func (this *generalMgr) HasGeneral(rid int, gid int) (*model.General, bool) {
	r, ok := this.GetByRId(rid)
	if ok {
		for _, v := range r {
			if v.Id == gid {
				return v, true
			}
		}
	}
	return nil, false
}

func (this *generalMgr) HasGenerals(rid int, gIds []int) ([]*model.General, bool) {
	gs := make([]*model.General, 0)
	for i := 0; i < len(gIds); i++ {
		g, ok := this.HasGeneral(rid, gIds[i])
		if ok {
			gs = append(gs, g)
		} else {
			return gs, false
		}
	}
	return gs, true
}

func (this *generalMgr) Count(rid int) int {
	gs, ok := this.GetByRId(rid)
	if ok {
		return len(gs)
	}
	return 0
}

// NewGeneral 创建并写库一个新武将。
func (this *generalMgr) NewGeneral(cfgId int, rid int, level int8) (*model.General, bool) {
	g, ok := model.NewGeneral(cfgId, rid, level)
	if !ok {
		return nil, false
	}
	c := coll(model.General{}.CollectionName())
	if c == nil {
		return nil, false
	}
	g.ServerId = serverID()
	g.Id = nextID(model.General{}.CollectionName())
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, g); err != nil {
		clog.Warn("[GMgr] NewGeneral insert err=%v", err)
		return nil, false
	}
	this.add(g)
	return g, true
}

func (this *generalMgr) GetOrCreateByRId(rid int) ([]*model.General, bool) {
	if r, ok := this.GetByRId(rid); ok {
		return r, true
	}
	return this.RandCreateGeneral(rid, 3)
}

func (this *generalMgr) RandCreateGeneral(rid int, nums int) ([]*model.General, bool) {
	gs := make([]*model.General, 0)
	for i := 0; i < nums; i++ {
		cfgId := general.General.Draw()
		g, ok := this.NewGeneral(cfgId, rid, 1)
		if !ok {
			return nil, false
		}
		gs = append(gs, g)
	}
	return gs, true
}

func (this *generalMgr) GetNPCGenerals(cfgIds []int, levels []int8) ([]model.General, bool) {
	if len(cfgIds) != len(levels) {
		return nil, false
	}
	gs, ok := this.GetByRId(0)
	if !ok {
		return nil, false
	}
	target := make([]model.General, 0, len(cfgIds))
	for i := 0; i < len(cfgIds); i++ {
		for _, g := range gs {
			if g.Level == levels[i] && g.CfgId == cfgIds[i] {
				target = append(target, *g)
				break
			}
		}
	}
	return target, true
}

func (this *generalMgr) GetDestroy(army *model.Army) int {
	destroy := 0
	for _, g := range army.Gens {
		if g != nil {
			destroy += g.GetDestroy()
		}
	}
	return destroy
}

func (this *generalMgr) PhysicalPowerIsEnough(army *model.Army, cost int) bool {
	for _, g := range army.Gens {
		if g == nil {
			continue
		}
		if g.PhysicalPower < cost {
			return false
		}
	}
	return true
}

func (this *generalMgr) TryUsePhysicalPower(army *model.Army, cost int) bool {
	if !this.PhysicalPowerIsEnough(army, cost) {
		return false
	}
	for _, g := range army.Gens {
		if g == nil {
			continue
		}
		g.PhysicalPower -= cost
		g.SyncExecute()
	}
	return true
}

// GeneralCfgInjector 注入到 model.GetGeneralCfg。
func GeneralCfgInjector(cfgId int) (model.GeneralCfg, bool) {
	c, ok := general.General.GMap[cfgId]
	if !ok {
		return model.GeneralCfg{}, false
	}
	return model.GeneralCfg{
		Force: c.Force, ForceGrow: c.ForceGrow,
		Strategy: c.Strategy, StrategyGrow: c.StrategyGrow,
		Defense: c.Defense, DefenseGrow: c.DefenseGrow,
		Speed: c.Speed, SpeedGrow: c.SpeedGrow,
		Destroy: c.Destroy, DestroyGrow: c.DestroyGrow,
		Camp: c.Camp, Star: c.Star, Arms: c.Arms,
	}, true
}

// NewGeneralBaseInit 注入到 model.NewGeneralBaseInit。
func NewGeneralBaseInit(cfgId int) (curArms int, star int8, plimit int, ok bool) {
	c, exists := general.General.GMap[cfgId]
	if !exists || len(c.Arms) == 0 {
		return 0, 0, 0, false
	}
	return c.Arms[0], c.Star, static_conf.Basic.General.PhysicalPowerLimit, true
}
