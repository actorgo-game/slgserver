package mgr

import (
	"sync"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
)

type armyMgr struct {
	mutex        sync.RWMutex
	armyById     map[int]*model.Army   // key: armyId
	armyByCityId map[int][]*model.Army // key: cityId
	armyByRId    map[int][]*model.Army // key: rid
}

var AMgr = &armyMgr{
	armyById:     make(map[int]*model.Army),
	armyByCityId: make(map[int][]*model.Army),
	armyByRId:    make(map[int][]*model.Army),
}

func (this *armyMgr) Load() {
	c := coll(model.Army{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[AMgr] Load find err=%v", err)
		return
	}
	var rows []*model.Army
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[AMgr] Load decode err=%v", err)
		return
	}
	for _, army := range rows {
		this.armyById[army.Id] = army
		army.CheckConscript()
		cid := army.CityId
		if _, ok := this.armyByCityId[cid]; !ok {
			this.armyByCityId[cid] = make([]*model.Army, 0)
		}
		this.armyByCityId[cid] = append(this.armyByCityId[cid], army)

		if _, ok := this.armyByRId[army.RId]; !ok {
			this.armyByRId[army.RId] = make([]*model.Army, 0)
		}
		this.armyByRId[army.RId] = append(this.armyByRId[army.RId], army)
		this.updateGenerals(army)
	}
}

func (this *armyMgr) insertOne(army *model.Army) {
	aid := army.Id
	cid := army.CityId
	this.armyById[aid] = army
	if _, ok := this.armyByCityId[cid]; !ok {
		this.armyByCityId[cid] = make([]*model.Army, 0)
	}
	this.armyByCityId[cid] = append(this.armyByCityId[cid], army)

	if _, ok := this.armyByRId[army.RId]; !ok {
		this.armyByRId[army.RId] = make([]*model.Army, 0)
	}
	this.armyByRId[army.RId] = append(this.armyByRId[army.RId], army)
	this.updateGenerals(army)
}

func (this *armyMgr) insertMutil(armys []*model.Army) {
	for _, v := range armys {
		this.insertOne(v)
	}
}

func (this *armyMgr) Get(aid int) (*model.Army, bool) {
	this.mutex.RLock()
	a, ok := this.armyById[aid]
	this.mutex.RUnlock()
	if ok {
		a.CheckConscript()
		return a, true
	}

	c := coll(model.Army{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	army := &model.Army{}
	if err := c.FindOne(cx, withServer(bson.M{"id": aid})).Decode(army); err != nil {
		return nil, false
	}
	this.mutex.Lock()
	this.insertOne(army)
	this.mutex.Unlock()
	return army, true
}

func (this *armyMgr) GetByCity(cid int) ([]*model.Army, bool) {
	this.mutex.RLock()
	as, ok := this.armyByCityId[cid]
	this.mutex.RUnlock()
	if ok {
		for _, a := range as {
			a.CheckConscript()
		}
		return as, true
	}

	c := coll(model.Army{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{"cityId": cid}))
	if err != nil {
		return nil, false
	}
	var m []*model.Army
	if err := cursor.All(cx, &m); err != nil {
		return nil, false
	}
	this.mutex.Lock()
	this.insertMutil(m)
	this.mutex.Unlock()
	return m, true
}

func (this *armyMgr) GetByCityOrder(cid int, order int8) (*model.Army, bool) {
	rs, ok := this.GetByCity(cid)
	if !ok {
		return nil, false
	}
	for _, r := range rs {
		if r.Order == order {
			return r, true
		}
	}
	return nil, false
}

func (this *armyMgr) GetByRId(rid int) ([]*model.Army, bool) {
	this.mutex.RLock()
	as, ok := this.armyByRId[rid]
	this.mutex.RUnlock()
	if ok {
		for _, a := range as {
			a.CheckConscript()
		}
	}
	return as, ok
}

func (this *armyMgr) BelongPosArmyCnt(rid int, x, y int) int {
	cnt := 0
	armys, ok := this.GetByRId(rid)
	if ok {
		for _, army := range armys {
			if army.FromX == x && army.FromY == y {
				cnt++
			} else if army.Cmd == model.ArmyCmdTransfer && army.ToX == x && army.ToY == y {
				cnt++
			}
		}
	}
	return cnt
}

func (this *armyMgr) GetOrCreate(rid int, cid int, order int8) (*model.Army, error) {
	this.mutex.RLock()
	armys, ok := this.armyByCityId[cid]
	this.mutex.RUnlock()
	if ok {
		for _, v := range armys {
			if v.Order == order {
				return v, nil
			}
		}
	}

	army := &model.Army{
		RId: rid, Order: order, CityId: cid,
		GeneralArray:       [model.ArmyGCnt]int{0, 0, 0},
		SoldierArray:       [model.ArmyGCnt]int{0, 0, 0},
		ConscriptCntArray:  [model.ArmyGCnt]int{0, 0, 0},
		ConscriptTimeArray: [model.ArmyGCnt]int64{0, 0, 0},
	}
	if city, ok := RCMgr.Get(cid); ok {
		army.FromX = city.X
		army.FromY = city.Y
		army.ToX = city.X
		army.ToY = city.Y
	}
	c := coll(model.Army{}.CollectionName())
	if c == nil {
		return nil, ErrDBNotReady
	}
	army.ServerId = serverID()
	army.Id = nextID(model.Army{}.CollectionName())
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, army); err != nil {
		return nil, err
	}
	this.mutex.Lock()
	this.insertOne(army)
	this.mutex.Unlock()
	return army, nil
}

func (this *armyMgr) GetSpeed(army *model.Army) int {
	speed := 100000
	for _, g := range army.Gens {
		if g != nil {
			s := g.GetSpeed()
			if s < speed {
				speed = s
			}
		}
	}

	camp := army.GetCamp()
	campAdds := []int{0}
	if camp > 0 {
		campAdds = RFMgr.GetAdditions(army.CityId, facility.TypeHanAddition-1+camp)
	}
	return speed + campAdds[0]
}

func (this *armyMgr) IsRepeat(rid int, cfgId int) bool {
	armys, ok := this.GetByRId(rid)
	if !ok {
		return true
	}
	for _, army := range armys {
		for _, g := range army.Gens {
			if g != nil && g.CfgId == cfgId && g.CityId != 0 {
				return false
			}
		}
	}
	return true
}

func (this *armyMgr) updateGenerals(armys ...*model.Army) {
	for _, army := range armys {
		for i, gid := range army.GeneralArray {
			if gid != 0 {
				g, _ := GMgr.GetByGId(gid)
				army.Gens[i] = g
			}
		}
	}
}

func (this *armyMgr) All() []*model.Army {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	armys := make([]*model.Army, 0, len(this.armyById))
	for _, army := range this.armyById {
		armys = append(armys, army)
	}
	return armys
}
