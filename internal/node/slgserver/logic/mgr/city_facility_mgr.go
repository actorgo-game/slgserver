package mgr

import (
	"sync"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
)

var RFMgr = facilityMgr{
	facilities:      make(map[int]*model.CityFacility),
	facilitiesByRId: make(map[int][]*model.CityFacility),
}

type facilityMgr struct {
	mutex           sync.RWMutex
	facilities      map[int]*model.CityFacility   // key: cityId
	facilitiesByRId map[int][]*model.CityFacility // key: rid
}

func (this *facilityMgr) Load() {
	c := coll(model.CityFacility{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()

	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[RFMgr] Load find err=%v", err)
		return
	}
	var rows []*model.CityFacility
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[RFMgr] Load decode err=%v", err)
		return
	}
	for _, cf := range rows {
		this.facilities[cf.CityId] = cf
		if _, ok := this.facilitiesByRId[cf.RId]; !ok {
			this.facilitiesByRId[cf.RId] = make([]*model.CityFacility, 0)
		}
		this.facilitiesByRId[cf.RId] = append(this.facilitiesByRId[cf.RId], cf)
	}
}

func (this *facilityMgr) GetByRId(rid int) ([]*model.CityFacility, bool) {
	this.mutex.RLock()
	r, ok := this.facilitiesByRId[rid]
	this.mutex.RUnlock()
	return r, ok
}

func (this *facilityMgr) Get(cid int) (*model.CityFacility, bool) {
	this.mutex.RLock()
	r, ok := this.facilities[cid]
	this.mutex.RUnlock()
	if ok {
		return r, true
	}

	c := coll(model.CityFacility{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	m := &model.CityFacility{}
	if err := c.FindOne(cx, withServer(bson.M{"cityId": cid})).Decode(m); err != nil {
		return nil, false
	}
	this.mutex.Lock()
	this.facilities[cid] = m
	this.mutex.Unlock()
	return m, true
}

func (this *facilityMgr) GetFacility(cid int, fType int8) (*model.Facility, bool) {
	cf, ok := this.Get(cid)
	if !ok {
		return nil, false
	}
	for i := range cf.Facilities {
		if cf.Facilities[i].Type == fType {
			return &cf.Facilities[i], true
		}
	}
	return nil, false
}

func (this *facilityMgr) GetFacilityLv(cid int, fType int8) int8 {
	if f, ok := this.GetFacility(cid, fType); ok {
		return f.GetLevel()
	}
	return 0
}

// GetAdditions 获取城内设施 additionType 的累计加成。
func (this *facilityMgr) GetAdditions(cid int, additionType ...int8) []int {
	cf, ok := this.Get(cid)
	ret := make([]int, len(additionType))
	if !ok {
		return ret
	}
	for i, at := range additionType {
		total := 0
		for _, f := range cf.Facility() {
			if f.GetLevel() > 0 {
				adds := facility.FConf.GetAdditions(f.Type)
				values := facility.FConf.GetValues(f.Type, f.GetLevel())
				for k, add := range adds {
					if add == at {
						total += values[k]
					}
				}
			}
		}
		ret[i] = total
	}
	return ret
}

// GetAndTryCreate 如果不存在尝试为该城创建一份默认设施列表。
func (this *facilityMgr) GetAndTryCreate(cid, rid int) (*model.CityFacility, bool) {
	if r, ok := this.Get(cid); ok {
		return r, true
	}
	if _, ok := RCMgr.Get(cid); !ok {
		clog.Warn("[RFMgr] GetAndTryCreate: cid not found cid=%d", cid)
		return nil, false
	}
	fs := make([]model.Facility, len(facility.FConf.List))
	for i, v := range facility.FConf.List {
		fs[i] = model.Facility{Type: v.Type, PrivateLevel: 0, Name: v.Name}
	}
	cf := &model.CityFacility{
		ServerId:   serverID(),
		Id:         nextID(model.CityFacility{}.CollectionName()),
		CityId:     cid,
		RId:        rid,
		Facilities: fs,
	}

	c := coll(model.CityFacility{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, cf); err != nil {
		clog.Warn("[RFMgr] insert facility err=%v", err)
		return nil, false
	}
	this.mutex.Lock()
	this.facilities[cid] = cf
	if _, ok := this.facilitiesByRId[rid]; !ok {
		this.facilitiesByRId[rid] = make([]*model.CityFacility, 0)
	}
	this.facilitiesByRId[rid] = append(this.facilitiesByRId[rid], cf)
	this.mutex.Unlock()
	return cf, true
}

func (this *facilityMgr) UpFacility(rid, cid int, fType int8) (*model.Facility, int32) {
	this.mutex.RLock()
	cf, ok := this.facilities[cid]
	this.mutex.RUnlock()
	if !ok {
		return nil, code.CityNotExist
	}

	for i := range cf.Facilities {
		fac := &cf.Facilities[i]
		if fac.Type != fType {
			continue
		}
		maxLevel := facility.FConf.MaxLevel(fType)
		if !fac.CanLV() {
			return nil, code.UpError
		}
		if fac.GetLevel() >= maxLevel {
			return nil, code.UpError
		}
		need, ok := facility.FConf.Need(fType, fac.GetLevel()+1)
		if !ok {
			return nil, code.UpError
		}
		if c := RResMgr.TryUseNeed(rid, *need); c != code.OK {
			return nil, c
		}
		fac.UpTime = time.Now().Unix()
		cf.SyncExecute()
		return fac, code.OK
	}
	return nil, code.UpError
}

func (this *facilityMgr) GetYield(rid int) model.Yield {
	cfs, ok := this.GetByRId(rid)
	var y model.Yield
	if !ok {
		return y
	}
	for _, cf := range cfs {
		for _, f := range cf.Facility() {
			if f.GetLevel() > 0 {
				values := facility.FConf.GetValues(f.Type, f.GetLevel())
				adds := facility.FConf.GetAdditions(f.Type)
				for i, at := range adds {
					switch at {
					case facility.TypeWood:
						y.Wood += values[i]
					case facility.TypeGrain:
						y.Grain += values[i]
					case facility.TypeIron:
						y.Iron += values[i]
					case facility.TypeStone:
						y.Stone += values[i]
					case facility.TypeTax:
						y.Gold += values[i]
					}
				}
			}
		}
	}
	return y
}

func (this *facilityMgr) GetDepotCapacity(rid int) int {
	cfs, ok := this.GetByRId(rid)
	limit := 0
	if !ok {
		return 0
	}
	for _, cf := range cfs {
		for _, f := range cf.Facility() {
			if f.GetLevel() > 0 {
				values := facility.FConf.GetValues(f.Type, f.GetLevel())
				adds := facility.FConf.GetAdditions(f.Type)
				for i, at := range adds {
					if at == facility.TypeWarehouseLimit {
						limit += values[i]
					}
				}
			}
		}
	}
	return limit
}

func (this *facilityMgr) GetCost(cid int) int8 {
	cf, ok := this.Get(cid)
	limit := 0
	if !ok {
		return 0
	}
	for _, f := range cf.Facility() {
		if f.GetLevel() > 0 {
			values := facility.FConf.GetValues(f.Type, f.GetLevel())
			adds := facility.FConf.GetAdditions(f.Type)
			for i, at := range adds {
				if at == facility.TypeCost {
					limit += values[i]
				}
			}
		}
	}
	return int8(limit)
}

func (this *facilityMgr) GetMaxDurable(cid int) int {
	cf, ok := this.Get(cid)
	limit := 0
	if !ok {
		return 0
	}
	for _, f := range cf.Facility() {
		if f.GetLevel() > 0 {
			values := facility.FConf.GetValues(f.Type, f.GetLevel())
			adds := facility.FConf.GetAdditions(f.Type)
			for i, at := range adds {
				if at == facility.TypeDurable {
					limit += values[i]
				}
			}
		}
	}
	return limit
}

func (this *facilityMgr) GetCityLV(cid int) int8 {
	return this.GetFacilityLv(cid, facility.Main)
}

// FacilityCostTime 注入到 model.FacilityCostTime。
func FacilityCostTime(t int8, nextLevel int8) int {
	return facility.FConf.CostTime(t, nextLevel)
}
