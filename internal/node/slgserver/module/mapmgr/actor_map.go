package mapmgr

import (
	"encoding/json"
	"os"
	"sync"

	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/entry"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/protocol"
)

const MapWidth = 200
const MapHeight = 200

type ActorMap struct {
	cactor.Base

	nationalMap map[int]*model.NationalMap
	builds      map[int]*model.MapRoleBuild
	posBuild    map[int]*model.MapRoleBuild
	posCity     map[int]*model.MapRoleCity

	buildEntry *entry.BuildEntry
	cityEntry  *entry.CityEntry

	posMu    sync.RWMutex
	posRoles map[int]map[int]bool
	ridPos   map[int][2]int
}

func NewActorMap() *ActorMap {
	return &ActorMap{
		nationalMap: make(map[int]*model.NationalMap),
		builds:      make(map[int]*model.MapRoleBuild),
		posBuild:    make(map[int]*model.MapRoleBuild),
		posCity:     make(map[int]*model.MapRoleCity),
		posRoles:    make(map[int]map[int]bool),
		ridPos:      make(map[int][2]int),
	}
}

func (p *ActorMap) AliasID() string {
	return "map"
}

func (p *ActorMap) OnInit() {
	p.initDB()
	p.loadNationalMap()
	p.loadBuilds()
	p.loadCities()

	p.Remote().Register("config", p.mapConfig)
	p.Remote().Register("scan", p.scan)
	p.Remote().Register("scanBlock", p.scanBlock)
	p.Remote().Register("giveUp", p.giveUp)
	p.Remote().Register("build", p.build)
	p.Remote().Register("upBuild", p.upBuild)
	p.Remote().Register("delBuild", p.delBuild)

	p.Remote().Register("occupyBuild", p.occupyBuild)
	p.Remote().Register("updatePos", p.updatePos)
	p.Remote().Register("getViewRIds", p.getViewRIds)

	clog.Info("[ActorMap] loaded: nationalMap=%d, builds=%d, cities=%d",
		len(p.nationalMap), len(p.builds), len(p.posCity))
}

func (p *ActorMap) initDB() {
	db := cmongo.Instance().GetDb("slg_db")
	if db == nil {
		return
	}
	sid := p.App().Settings().GetInt("server_id", 1)
	p.buildEntry = entry.NewBuildEntry(db.Collection("builds"), sid)
	p.cityEntry = entry.NewCityEntry(db.Collection("cities"), sid)
}

func (p *ActorMap) loadNationalMap() {
	jsonPath := p.App().Settings().GetString("json_data", "config/data/")
	data, err := os.ReadFile(jsonPath + "map_build.json")
	if err != nil {
		clog.Warn("[ActorMap] load map_build.json fail: %v", err)
		return
	}
	var items []model.NationalMap
	if err := json.Unmarshal(data, &items); err != nil {
		clog.Warn("[ActorMap] parse map_build.json fail: %v", err)
		return
	}
	for i := range items {
		posId := items[i].X*MapWidth + items[i].Y
		items[i].MId = posId
		p.nationalMap[posId] = &items[i]
	}
}

func (p *ActorMap) loadBuilds() {
	if p.buildEntry == nil {
		return
	}
	builds, err := p.buildEntry.FindAll()
	if err != nil {
		clog.Warn("[ActorMap] load builds fail: %v", err)
		return
	}
	for _, b := range builds {
		p.builds[b.Id] = b
		posId := b.X*MapWidth + b.Y
		p.posBuild[posId] = b
	}
}

func (p *ActorMap) loadCities() {
	if p.cityEntry == nil {
		return
	}
	cities, err := p.cityEntry.FindAll()
	if err != nil {
		clog.Warn("[ActorMap] load cities fail: %v", err)
		return
	}
	for _, c := range cities {
		posId := c.X*MapWidth + c.Y
		p.posCity[posId] = c
	}
}

func posToId(x, y int) int {
	return x*MapWidth + y
}

// --- Position tracking ---

func (p *ActorMap) updatePos(req *protocol.UpPositionReq) {
}

func (p *ActorMap) UpdateRolePos(rid, x, y int) {
	p.posMu.Lock()
	defer p.posMu.Unlock()

	if old, ok := p.ridPos[rid]; ok {
		oldPosId := posToId(old[0], old[1])
		if m, ok2 := p.posRoles[oldPosId]; ok2 {
			delete(m, rid)
		}
	}

	newPosId := posToId(x, y)
	if _, ok := p.posRoles[newPosId]; !ok {
		p.posRoles[newPosId] = make(map[int]bool)
	}
	p.posRoles[newPosId][rid] = true
	p.ridPos[rid] = [2]int{x, y}
}

type viewRIdsReq struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type viewRIdsRsp struct {
	RIds []int `json:"rids"`
}

func (p *ActorMap) getViewRIds(req *viewRIdsReq) (*viewRIdsRsp, int32) {
	p.posMu.RLock()
	defer p.posMu.RUnlock()

	rids := make([]int, 0)
	for i := req.X - req.Width; i <= req.X+req.Width; i++ {
		for j := req.Y - req.Height; j <= req.Y+req.Height; j++ {
			posId := posToId(i, j)
			if m, ok := p.posRoles[posId]; ok {
				for rid := range m {
					rids = append(rids, rid)
				}
			}
		}
	}
	return &viewRIdsRsp{RIds: rids}, code.OK
}

// --- Remote handlers ---

func (p *ActorMap) mapConfig(_ *protocol.ConfigReq) (*protocol.ConfigRsp, int32) {
	confs := make([]protocol.Conf, 0, len(p.nationalMap))
	for _, nm := range p.nationalMap {
		confs = append(confs, protocol.Conf{
			Type:  nm.Type,
			Level: nm.Level,
		})
	}
	return &protocol.ConfigRsp{Confs: confs}, code.OK
}

func (p *ActorMap) scan(req *protocol.ScanReq) (*protocol.ScanRsp, int32) {
	rsp := &protocol.ScanRsp{
		MRBuilds: make([]protocol.MapRoleBuild, 0),
		MCBuilds: make([]protocol.MapRoleCity, 0),
		Armys:    make([]protocol.Army, 0),
	}

	defaultRadius := 10
	for x := req.X - defaultRadius; x <= req.X+defaultRadius; x++ {
		for y := req.Y - defaultRadius; y <= req.Y+defaultRadius; y++ {
			posId := posToId(x, y)

			if b, ok := p.posBuild[posId]; ok {
				rsp.MRBuilds = append(rsp.MRBuilds, toBuildProto(b))
			}
			if c, ok := p.posCity[posId]; ok {
				rsp.MCBuilds = append(rsp.MCBuilds, toCityProto(c))
			}
		}
	}

	return rsp, code.OK
}

func (p *ActorMap) scanBlock(req *protocol.ScanBlockReq) (*protocol.ScanRsp, int32) {
	rsp := &protocol.ScanRsp{
		MRBuilds: make([]protocol.MapRoleBuild, 0),
		MCBuilds: make([]protocol.MapRoleCity, 0),
		Armys:    make([]protocol.Army, 0),
	}

	halfLen := req.Length / 2
	cx := req.X + halfLen
	cy := req.Y + halfLen
	for x := cx - halfLen; x <= cx+halfLen; x++ {
		for y := cy - halfLen; y <= cy+halfLen; y++ {
			posId := posToId(x, y)

			if b, ok := p.posBuild[posId]; ok {
				rsp.MRBuilds = append(rsp.MRBuilds, toBuildProto(b))
			}
			if c, ok := p.posCity[posId]; ok {
				rsp.MCBuilds = append(rsp.MCBuilds, toCityProto(c))
			}
		}
	}

	return rsp, code.OK
}

func (p *ActorMap) giveUp(req *protocol.GiveUpReq) (*protocol.GiveUpRsp, int32) {
	posId := posToId(req.X, req.Y)
	b, ok := p.posBuild[posId]
	if !ok {
		return nil, code.InvalidParam
	}
	delete(p.builds, b.Id)
	delete(p.posBuild, posId)
	if p.buildEntry != nil {
		_ = p.buildEntry.Delete(b.Id)
	}
	return &protocol.GiveUpRsp{X: req.X, Y: req.Y}, code.OK
}

func (p *ActorMap) build(req *protocol.BuildReq) (*protocol.BuildRsp, int32) {
	posId := posToId(req.X, req.Y)
	if _, ok := p.posBuild[posId]; ok {
		return nil, code.CanNotBuildNew
	}
	if _, ok := p.posCity[posId]; ok {
		return nil, code.CanNotBuildNew
	}

	b := &model.MapRoleBuild{
		Id:   posId,
		Type: req.Type,
		X:    req.X,
		Y:    req.Y,
	}
	p.builds[b.Id] = b
	p.posBuild[posId] = b
	if p.buildEntry != nil {
		_ = p.buildEntry.Insert(b)
	}
	return &protocol.BuildRsp{X: req.X, Y: req.Y, Type: req.Type}, code.OK
}

func (p *ActorMap) upBuild(req *protocol.UpBuildReq) (*protocol.UpBuildRsp, int32) {
	posId := posToId(req.X, req.Y)
	b, ok := p.posBuild[posId]
	if !ok {
		return nil, code.InvalidParam
	}
	b.Level++
	if p.buildEntry != nil {
		_ = p.buildEntry.Update(b)
	}
	return &protocol.UpBuildRsp{X: req.X, Y: req.Y, Build: toBuildProto(b)}, code.OK
}

func (p *ActorMap) delBuild(req *protocol.DelBuildReq) (*protocol.DelBuildRsp, int32) {
	posId := posToId(req.X, req.Y)
	b, ok := p.posBuild[posId]
	if !ok {
		return nil, code.InvalidParam
	}
	bp := toBuildProto(b)
	delete(p.builds, b.Id)
	delete(p.posBuild, posId)
	if p.buildEntry != nil {
		_ = p.buildEntry.Delete(b.Id)
	}
	return &protocol.DelBuildRsp{X: req.X, Y: req.Y, Build: bp}, code.OK
}

func (p *ActorMap) occupyBuild(req *model.MapRoleBuild) int32 {
	if b, ok := p.builds[req.Id]; ok {
		b.RId = req.RId
		b.OccupyTime = req.OccupyTime
		if p.buildEntry != nil {
			_ = p.buildEntry.Update(b)
		}
		return code.OK
	}
	return code.InvalidParam
}

// --- helpers ---

func toBuildProto(b *model.MapRoleBuild) protocol.MapRoleBuild {
	return protocol.MapRoleBuild{
		RId: b.RId, Name: b.Name, Type: b.Type, Level: b.Level,
		OPLevel: b.OPLevel, X: b.X, Y: b.Y,
		CurDurable: b.CurDurable, MaxDurable: b.MaxDurable,
		OccupyTime: b.OccupyTime.Unix(), EndTime: b.EndTime.Unix(),
		GiveUpTime: b.GiveUpTime,
	}
}

func toCityProto(c *model.MapRoleCity) protocol.MapRoleCity {
	return protocol.MapRoleCity{
		CityId: c.CityId, RId: c.RId, Name: c.Name,
		X: c.X, Y: c.Y, IsMain: c.IsMain,
		CurDurable: c.CurDurable, MaxDurable: c.MaxDurable,
		OccupyTime: c.OccupyTime.Unix(),
	}
}
