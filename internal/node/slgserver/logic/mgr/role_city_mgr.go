package mgr

import (
	"sync"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/global"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/util"
)

type roleCityMgr struct {
	mutex    sync.RWMutex
	dbCity   map[int]*model.MapRoleCity   // key: cityId
	posCity  map[int]*model.MapRoleCity   // key: posId
	roleCity map[int][]*model.MapRoleCity // key: rid
}

var RCMgr = &roleCityMgr{
	dbCity:   make(map[int]*model.MapRoleCity),
	posCity:  make(map[int]*model.MapRoleCity),
	roleCity: make(map[int][]*model.MapRoleCity),
}

// GetCityCost 注入到 model.GetCityCost。
func GetCityCost(cid int) int8 {
	return RFMgr.GetCost(cid) + static_conf.Basic.City.Cost
}

// GetMaxDurable 注入到 model.GetMaxDurable。
func GetMaxDurable(cid int) int {
	return RFMgr.GetMaxDurable(cid) + static_conf.Basic.City.Durable
}

// GetCityLV 注入到 model.GetCityLv。
func GetCityLV(cid int) int8 {
	return RFMgr.GetCityLV(cid)
}

func (this *roleCityMgr) Load() {
	c := coll(model.MapRoleCity{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[RCMgr] Load find err=%v", err)
		return
	}
	var rows []*model.MapRoleCity
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[RCMgr] Load decode err=%v", err)
		return
	}

	for _, v := range rows {
		this.dbCity[v.CityId] = v
		posId := global.ToPosition(v.X, v.Y)
		this.posCity[posId] = v
		if _, ok := this.roleCity[v.RId]; !ok {
			this.roleCity[v.RId] = make([]*model.MapRoleCity, 0)
		}
		this.roleCity[v.RId] = append(this.roleCity[v.RId], v)
	}

	go this.running()
}

func (this *roleCityMgr) running() {
	for {
		t := static_conf.Basic.City.RecoveryTime
		time.Sleep(time.Duration(t) * time.Second)
		this.mutex.RLock()
		for _, city := range this.dbCity {
			if city.CurDurable < GetMaxDurable(city.CityId) {
				city.DurableChange(100)
				city.SyncExecute()
			}
		}
		this.mutex.RUnlock()
	}
}

func (this *roleCityMgr) IsEmpty(x, y int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := global.ToPosition(x, y)
	_, ok := this.posCity[posId]
	return !ok
}

func (this *roleCityMgr) PositionCity(x, y int) (*model.MapRoleCity, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := global.ToPosition(x, y)
	c, ok := this.posCity[posId]
	return c, ok
}

func (this *roleCityMgr) Add(city *model.MapRoleCity) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.dbCity[city.CityId] = city
	this.posCity[global.ToPosition(city.X, city.Y)] = city
	if _, ok := this.roleCity[city.RId]; !ok {
		this.roleCity[city.RId] = make([]*model.MapRoleCity, 0)
	}
	this.roleCity[city.RId] = append(this.roleCity[city.RId], city)
}

// Insert 写入新城市到 mongo，自增 cityId，并加入缓存。
func (this *roleCityMgr) Insert(city *model.MapRoleCity) error {
	c := coll(model.MapRoleCity{}.CollectionName())
	if c == nil {
		return ErrDBNotReady
	}
	if city.CityId == 0 {
		city.CityId = nextID(model.MapRoleCity{}.CollectionName())
	}
	city.ServerId = serverID()
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, city); err != nil {
		return err
	}
	this.Add(city)
	return nil
}

func (this *roleCityMgr) Scan(x, y int) []*model.MapRoleCity {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(40, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(40, y+ScanHeight)

	cb := make([]*model.MapRoleCity, 0)
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			if v, ok := this.posCity[global.ToPosition(i, j)]; ok {
				cb = append(cb, v)
			}
		}
	}
	return cb
}

func (this *roleCityMgr) ScanBlock(x, y, length int) []*model.MapRoleCity {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)

	cb := make([]*model.MapRoleCity, 0)
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			if v, ok := this.posCity[global.ToPosition(i, j)]; ok {
				cb = append(cb, v)
			}
		}
	}
	return cb
}

func (this *roleCityMgr) GetByRId(rid int) ([]*model.MapRoleCity, bool) {
	this.mutex.RLock()
	r, ok := this.roleCity[rid]
	this.mutex.RUnlock()
	return r, ok
}

func (this *roleCityMgr) GetMainCity(rid int) (*model.MapRoleCity, bool) {
	citys, ok := this.GetByRId(rid)
	if !ok {
		return nil, false
	}
	for _, city := range citys {
		if city.IsMain == 1 {
			return city, true
		}
	}
	return nil, false
}

func (this *roleCityMgr) Get(cid int) (*model.MapRoleCity, bool) {
	this.mutex.RLock()
	r, ok := this.dbCity[cid]
	this.mutex.RUnlock()
	if ok {
		return r, true
	}

	c := coll(model.MapRoleCity{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	m := &model.MapRoleCity{}
	if err := c.FindOne(cx, withServer(bson.M{"cityId": cid})).Decode(m); err != nil {
		return nil, false
	}
	this.mutex.Lock()
	this.dbCity[cid] = m
	this.mutex.Unlock()
	return m, true
}
