package mgr

import (
	"sync"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
	"github.com/llr104/slgserver/internal/util"
)

type roleResMgr struct {
	mutex    sync.RWMutex
	rolesRes map[int]*model.RoleRes // key: rid
}

var RResMgr = &roleResMgr{
	rolesRes: make(map[int]*model.RoleRes),
}

// GetYield 角色总产量 = 建筑产出 + 城市设施产出 + 角色基础产出。
// 与原 mgr.GetYield 同语义。注入到 model.GetYield。
func GetYield(rid int) model.Yield {
	by := RBMgr.GetYield(rid)
	cy := RFMgr.GetYield(rid)
	var y model.Yield

	y.Gold = by.Gold + cy.Gold + static_conf.Basic.Role.GoldYield
	y.Stone = by.Stone + cy.Stone + static_conf.Basic.Role.StoneYield
	y.Iron = by.Iron + cy.Iron + static_conf.Basic.Role.IronYield
	y.Grain = by.Grain + cy.Grain + static_conf.Basic.Role.GrainYield
	y.Wood = by.Wood + cy.Wood + static_conf.Basic.Role.WoodYield
	return y
}

// GetDepotCapacity 仓库总容量。注入到 model.GetDepotCapacity。
func GetDepotCapacity(rid int) int {
	return RFMgr.GetDepotCapacity(rid) + static_conf.Basic.Role.DepotCapacity
}

func (this *roleResMgr) Load() {
	c := coll(model.RoleRes{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[RResMgr] Load find err=%v", err)
		return
	}
	var rows []*model.RoleRes
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[RResMgr] Load decode err=%v", err)
		return
	}
	for _, v := range rows {
		this.rolesRes[v.RId] = v
	}

	go this.produce()
}

func (this *roleResMgr) Get(rid int) (*model.RoleRes, bool) {
	this.mutex.RLock()
	r, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	if ok {
		return r, true
	}

	c := coll(model.RoleRes{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	m := &model.RoleRes{}
	if err := c.FindOne(cx, withServer(bson.M{"rid": rid})).Decode(m); err != nil {
		return nil, false
	}
	this.mutex.Lock()
	this.rolesRes[rid] = m
	this.mutex.Unlock()
	return m, true
}

func (this *roleResMgr) Add(res *model.RoleRes) {
	this.mutex.Lock()
	this.rolesRes[res.RId] = res
	this.mutex.Unlock()
}

// Insert 写入新的资源记录到 mongo（enterServer 时调用）。
func (this *roleResMgr) Insert(res *model.RoleRes) error {
	c := coll(model.RoleRes{}.CollectionName())
	if c == nil {
		return ErrDBNotReady
	}
	if res.Id == 0 {
		res.Id = nextID(model.RoleRes{}.CollectionName())
	}
	res.ServerId = serverID()
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, res); err != nil {
		return err
	}
	this.Add(res)
	return nil
}

func (this *roleResMgr) TryUseNeed(rid int, need facility.NeedRes) int32 {
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if !ok {
		return code.RoleNotExist
	}

	if need.Decree <= rr.Decree && need.Grain <= rr.Grain &&
		need.Stone <= rr.Stone && need.Wood <= rr.Wood &&
		need.Iron <= rr.Iron && need.Gold <= rr.Gold {
		rr.Decree -= need.Decree
		rr.Iron -= need.Iron
		rr.Wood -= need.Wood
		rr.Stone -= need.Stone
		rr.Grain -= need.Grain
		rr.Gold -= need.Gold

		rr.SyncExecute()
		return code.OK
	}
	if need.Decree > rr.Decree {
		return code.DecreeNotEnough
	}
	return code.ResNotEnough
}

func (this *roleResMgr) DecreeIsEnough(rid int, cost int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if rr, ok := this.rolesRes[rid]; ok {
		return rr.Decree >= cost
	}
	return false
}

func (this *roleResMgr) TryUseDecree(rid int, decree int) bool {
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	if !ok {
		return false
	}
	if rr.Decree >= decree {
		rr.Decree -= decree
		rr.SyncExecute()
		return true
	}
	return false
}

func (this *roleResMgr) GoldIsEnough(rid int, cost int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if rr, ok := this.rolesRes[rid]; ok {
		return rr.Gold >= cost
	}
	return false
}

func (this *roleResMgr) TryUseGold(rid int, gold int) bool {
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	if !ok {
		return false
	}
	if rr.Gold >= gold {
		rr.Gold -= gold
		rr.SyncExecute()
		return true
	}
	return false
}

// produce 周期性给所有玩家增加资源。结构与原 slgserver 一致：
//
//	每 RecoveryTime 秒执行一次，每次 +yield/6（即每秒一次的"小步长"）。
func (this *roleResMgr) produce() {
	index := 1
	for {
		t := static_conf.Basic.Role.RecoveryTime
		time.Sleep(time.Duration(t) * time.Second)
		this.mutex.RLock()
		for _, v := range this.rolesRes {
			capacity := GetDepotCapacity(v.RId)
			y := GetYield(v.RId)

			if v.Wood < capacity {
				v.Wood = util.MinInt(v.Wood+y.Wood/6, capacity)
			}
			if v.Iron < capacity {
				v.Iron = util.MinInt(v.Iron+y.Iron/6, capacity)
			}
			if v.Stone < capacity {
				v.Stone = util.MinInt(v.Stone+y.Stone/6, capacity)
			}
			if v.Grain < capacity {
				v.Grain = util.MinInt(v.Grain+y.Grain/6, capacity)
			}
			if v.Gold < capacity {
				v.Gold = util.MinInt(v.Gold+y.Gold/6, capacity)
			}

			if index%6 == 0 {
				if v.Decree < static_conf.Basic.Role.DecreeLimit {
					v.Decree += 1
				}
			}
			v.SyncExecute()
		}
		index++
		this.mutex.RUnlock()
	}
}
