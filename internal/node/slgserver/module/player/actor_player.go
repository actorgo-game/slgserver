package player

import (
	"fmt"
	"strconv"
	"time"

	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/entry"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/module/online"
	"github.com/llr104/slgserver/internal/protocol"
	"github.com/llr104/slgserver/internal/util"
)

type actorPlayer struct {
	cactor.Base

	isOnline bool
	rid      int
	uid      int64
	serverId int

	role       *model.Role
	cities     []*model.MapRoleCity
	generals   []*model.General
	armies     []*model.Army
	skills     []*model.Skill
	facilities []*model.CityFacility
	resources  *model.RoleRes
	attribute  *model.RoleAttribute

	roleEntry      *entry.RoleEntry
	generalEntry   *entry.GeneralEntry
	armyEntry      *entry.ArmyEntry
	cityEntry      *entry.CityEntry
	buildEntry     *entry.BuildEntry
	facilityEntry  *entry.FacilityEntry
	resourceEntry  *entry.ResourceEntry
	attributeEntry *entry.AttributeEntry
	skillEntry     *entry.SkillEntry
	warReportEntry *entry.WarReportEntry
}

func NewActorPlayer() *actorPlayer {
	return &actorPlayer{}
}

func (p *actorPlayer) OnInit() {
	clog.Debug("[actorPlayer] init: path=%s", p.PathString())

	p.initDB()

	p.Remote().Register("sessionClose", p.sessionClose)
	p.Remote().Register("updateArmyResult", p.updateArmyResult)

	p.Remote().Register("roleList", p.roleList)
	p.Remote().Register("create", p.createRole)
	p.Remote().Register("enterServer", p.enterServer)
	p.Remote().Register("myCity", p.myCity)
	p.Remote().Register("myRoleRes", p.myRoleRes)
	p.Remote().Register("myRoleBuild", p.myRoleBuild)
	p.Remote().Register("myProperty", p.myProperty)
	p.Remote().Register("upPosition", p.upPosition)
	p.Remote().Register("posTagList", p.posTagList)
	p.Remote().Register("opPosTag", p.opPosTag)

	p.Remote().Register("myGenerals", p.myGenerals)
	p.Remote().Register("drawGeneral", p.drawGeneral)
	p.Remote().Register("composeGeneral", p.composeGeneral)
	p.Remote().Register("addPrGeneral", p.addPrGeneral)
	p.Remote().Register("convert", p.convertGeneral)
	p.Remote().Register("upSkill", p.upSkill)
	p.Remote().Register("downSkill", p.downSkill)
	p.Remote().Register("lvSkill", p.lvSkill)

	p.Remote().Register("myList", p.armyList)
	p.Remote().Register("myOne", p.armyOne)
	p.Remote().Register("dispose", p.dispose)
	p.Remote().Register("conscript", p.conscript)
	p.Remote().Register("assign", p.assignArmy)

	p.Remote().Register("facilities", p.facilityList)
	p.Remote().Register("upFacility", p.upFacility)

	p.Remote().Register("collect", p.collect)
	p.Remote().Register("openCollect", p.openCollect)
	p.Remote().Register("transform", p.transform)

	p.Remote().Register("list", p.skillList)

	p.Remote().Register("report", p.warReport)
	p.Remote().Register("read", p.warRead)

	p.Timer().Add(60*time.Second, p.timerSave)
	p.Timer().Add(60*time.Second, p.timerProduceRes)
}

func (p *actorPlayer) initDB() {
	db := cmongo.Instance().GetDb("slg_db")
	if db == nil {
		clog.Error("[actorPlayer] slg_db not found")
		return
	}

	sid := p.App().Settings().GetInt("server_id", 1)
	p.serverId = sid

	p.roleEntry = entry.NewRoleEntry(db.Collection("roles"), sid)
	p.generalEntry = entry.NewGeneralEntry(db.Collection("generals"), sid)
	p.armyEntry = entry.NewArmyEntry(db.Collection("armies"), sid)
	p.cityEntry = entry.NewCityEntry(db.Collection("cities"), sid)
	p.buildEntry = entry.NewBuildEntry(db.Collection("builds"), sid)
	p.facilityEntry = entry.NewFacilityEntry(db.Collection("city_facilities"), sid)
	p.resourceEntry = entry.NewResourceEntry(db.Collection("role_resources"), sid)
	p.attributeEntry = entry.NewAttributeEntry(db.Collection("role_attributes"), sid)
	p.skillEntry = entry.NewSkillEntry(db.Collection("skills"), sid)
	p.warReportEntry = entry.NewWarReportEntry(db.Collection("war_reports"), sid)
}

func (p *actorPlayer) OnStop() {
	p.saveAll()
	clog.Debug("[actorPlayer] stop: rid=%d", p.rid)
}

func (p *actorPlayer) getUid() int64 {
	uid, _ := strconv.ParseInt(p.ActorID(), 10, 64)
	return uid
}

func (p *actorPlayer) sessionClose(_ *struct{}) int32 {
	online.UnBindPlayer(p.uid)
	p.isOnline = false
	p.saveAll()
	p.Exit()
	clog.Debug("[actorPlayer] session closed: uid=%d, rid=%d", p.uid, p.rid)
	return code.OK
}

func (p *actorPlayer) loadPlayerData(rid int) bool {
	role, err := p.roleEntry.FindByRId(rid)
	if err != nil {
		return false
	}
	p.role = role
	p.rid = role.RId
	p.uid = int64(role.UId)

	p.cities, _ = p.cityEntry.FindByRId(rid)
	p.generals, _ = p.generalEntry.FindByRId(rid)
	p.armies, _ = p.armyEntry.FindByRId(rid)
	p.skills, _ = p.skillEntry.FindByRId(rid)
	p.resources, _ = p.resourceEntry.FindByRId(rid)
	p.attribute, _ = p.attributeEntry.FindByRId(rid)

	if p.cities == nil {
		p.cities = []*model.MapRoleCity{}
	}
	if p.generals == nil {
		p.generals = []*model.General{}
	}
	if p.armies == nil {
		p.armies = []*model.Army{}
	}
	if p.skills == nil {
		p.skills = []*model.Skill{}
	}
	if p.resources == nil {
		p.resources = &model.RoleRes{RId: rid}
	}
	if p.attribute == nil {
		p.attribute = &model.RoleAttribute{RId: rid}
	}

	for _, city := range p.cities {
		facs, _ := p.facilityEntry.FindByRId(rid)
		p.facilities = facs
		_ = city
		break
	}
	if p.facilities == nil {
		p.facilities = []*model.CityFacility{}
	}

	return true
}

func (p *actorPlayer) saveAll() {
	if p.role != nil {
		_ = p.roleEntry.Update(p.role)
	}
	if p.resources != nil {
		_ = p.resourceEntry.Upsert(p.resources)
	}
	if p.attribute != nil {
		_ = p.attributeEntry.Upsert(p.attribute)
	}
	for _, g := range p.generals {
		_ = p.generalEntry.Update(g)
	}
	for _, a := range p.armies {
		_ = p.armyEntry.Update(a)
	}
	for _, f := range p.facilities {
		_ = p.facilityEntry.Upsert(f)
	}
}

func (p *actorPlayer) timerSave() {
	if p.isOnline {
		p.saveAll()
	}
}

func (p *actorPlayer) timerProduceRes() {
	if !p.isOnline || p.resources == nil {
		return
	}
	p.resources.Wood += 100
	p.resources.Iron += 100
	p.resources.Stone += 100
	p.resources.Grain += 100
	p.resources.Gold += 10
}

// --- Role handlers ---

func (p *actorPlayer) roleList(_ *protocol.RoleListReq) (*protocol.RoleListRsp, int32) {
	uid := p.getUid()
	rsp := &protocol.RoleListRsp{Roles: []protocol.Role{}}

	role, err := p.roleEntry.FindByUId(int(uid))
	if err == nil && role != nil {
		rsp.Roles = append(rsp.Roles, toRole(role))
	}

	return rsp, code.OK
}

func (p *actorPlayer) createRole(req *protocol.CreateRoleReq) (*protocol.CreateRoleRsp, int32) {
	uid := p.getUid()

	existing, _ := p.roleEntry.FindByUId(int(uid))
	if existing != nil {
		return nil, code.RoleAlreadyCreate
	}

	rid := int(time.Now().UnixNano() % 1000000)
	role := &model.Role{
		RId:       rid,
		UId:       int(uid),
		NickName:  req.NickName,
		HeadId:    int(req.HeadId),
		Sex:       req.Sex,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	if err := p.roleEntry.Insert(role); err != nil {
		clog.Warn("[createRole] insert fail: %v", err)
		return nil, code.DBError
	}

	city := &model.MapRoleCity{
		CityId:     rid*10 + 1,
		RId:        rid,
		Name:       req.NickName + " City",
		X:          rid % 200,
		Y:          rid / 200 % 200,
		IsMain:     true,
		CurDurable: 1000,
		MaxDurable: 1000,
		CreatedAt:  time.Now(),
	}
	_ = p.cityEntry.Insert(city)

	res := &model.RoleRes{
		RId:    rid,
		Wood:   5000,
		Iron:   5000,
		Stone:  5000,
		Grain:  5000,
		Gold:   1000,
		Decree: 20,
	}
	_ = p.resourceEntry.Upsert(res)

	attr := &model.RoleAttribute{RId: rid}
	_ = p.attributeEntry.Upsert(attr)

	for i := 0; i < 5; i++ {
		army := &model.Army{
			Id:     rid*10 + i + 1,
			RId:    rid,
			CityId: city.CityId,
			Order:  int8(i + 1),
			Cmd:    model.ArmyCmdIdle,
		}
		_ = p.armyEntry.Insert(army)
	}

	return &protocol.CreateRoleRsp{Role: toRole(role)}, code.OK
}

func (p *actorPlayer) enterServer(req *protocol.EnterServerReq) (*protocol.EnterServerRsp, int32) {
	uid := p.getUid()

	var rid int
	sess, err := util.ParseSession(req.Session)
	if err == nil && sess.IsValid() {
		rid = sess.Id
	}

	if rid <= 0 {
		role, findErr := p.roleEntry.FindByUId(int(uid))
		if findErr != nil || role == nil {
			return nil, code.RoleNotExist
		}
		rid = role.RId
	}

	if !p.loadPlayerData(rid) {
		return nil, code.RoleNotExist
	}

	p.role.LoginTime = time.Now()
	p.uid = uid
	p.isOnline = true

	online.BindPlayer(p.rid, p.uid, p.PathString())

	token := util.NewSession(int(uid), time.Now()).String()

	rsp := &protocol.EnterServerRsp{
		Role:    toRole(p.role),
		RoleRes: toRoleRes(p.resources),
		Time:    time.Now().Unix(),
		Token:   token,
	}
	return rsp, code.OK
}

func (p *actorPlayer) myCity(_ *protocol.MyCityReq) (*protocol.MyCityRsp, int32) {
	rsp := &protocol.MyCityRsp{Citys: make([]protocol.MapRoleCity, 0, len(p.cities))}
	for _, c := range p.cities {
		rsp.Citys = append(rsp.Citys, toCity(c))
	}
	return rsp, code.OK
}

func (p *actorPlayer) myRoleRes(_ *protocol.MyRoleResReq) (*protocol.MyRoleResRsp, int32) {
	return &protocol.MyRoleResRsp{RoleRes: toRoleRes(p.resources)}, code.OK
}

func (p *actorPlayer) myRoleBuild(_ *protocol.MyRoleBuildReq) (*protocol.MyRoleBuildRsp, int32) {
	builds, _ := p.buildEntry.FindByRId(p.rid)
	rsp := &protocol.MyRoleBuildRsp{MRBuilds: make([]protocol.MapRoleBuild, 0)}
	for _, b := range builds {
		rsp.MRBuilds = append(rsp.MRBuilds, toBuild(b))
	}
	return rsp, code.OK
}

func (p *actorPlayer) myProperty(_ *protocol.MyRolePropertyReq) (*protocol.MyRolePropertyRsp, int32) {
	rsp := &protocol.MyRolePropertyRsp{
		RoleRes:  toRoleRes(p.resources),
		Citys:    make([]protocol.MapRoleCity, 0, len(p.cities)),
		Generals: make([]protocol.General, 0, len(p.generals)),
		Armys:    make([]protocol.Army, 0, len(p.armies)),
	}
	for _, c := range p.cities {
		rsp.Citys = append(rsp.Citys, toCity(c))
	}
	for _, g := range p.generals {
		rsp.Generals = append(rsp.Generals, toGeneral(g))
	}
	for _, a := range p.armies {
		rsp.Armys = append(rsp.Armys, toArmy(a))
	}
	builds, _ := p.buildEntry.FindByRId(p.rid)
	rsp.MRBuilds = make([]protocol.MapRoleBuild, 0, len(builds))
	for _, b := range builds {
		rsp.MRBuilds = append(rsp.MRBuilds, toBuild(b))
	}
	return rsp, code.OK
}

func (p *actorPlayer) upPosition(req *protocol.UpPositionReq) (*protocol.UpPositionRsp, int32) {
	return &protocol.UpPositionRsp{X: req.X, Y: req.Y}, code.OK
}

func (p *actorPlayer) posTagList(_ *protocol.PosTagListReq) (*protocol.PosTagListRsp, int32) {
	rsp := &protocol.PosTagListRsp{PosTags: make([]protocol.PosTag, 0)}
	if p.attribute != nil {
		for _, t := range p.attribute.PosTags {
			rsp.PosTags = append(rsp.PosTags, protocol.PosTag{X: t.X, Y: t.Y, Name: t.Name})
		}
	}
	return rsp, code.OK
}

func (p *actorPlayer) opPosTag(req *protocol.PosTagReq) (*protocol.PosTagRsp, int32) {
	if p.attribute == nil {
		p.attribute = &model.RoleAttribute{RId: p.rid}
	}

	found := false
	for i, t := range p.attribute.PosTags {
		if t.X == req.X && t.Y == req.Y {
			p.attribute.PosTags = append(p.attribute.PosTags[:i], p.attribute.PosTags[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		if len(p.attribute.PosTags) >= 20 {
			return nil, code.OutPosTagLimit
		}
		p.attribute.PosTags = append(p.attribute.PosTags, model.PosTag{X: req.X, Y: req.Y, Name: req.Name})
	}
	_ = p.attributeEntry.Upsert(p.attribute)

	return &protocol.PosTagRsp{Type: req.Type, X: req.X, Y: req.Y, Name: req.Name}, code.OK
}

// --- General handlers ---

func (p *actorPlayer) myGenerals(_ *protocol.MyGeneralReq) (*protocol.MyGeneralRsp, int32) {
	rsp := &protocol.MyGeneralRsp{Generals: make([]protocol.General, 0, len(p.generals))}
	for _, g := range p.generals {
		rsp.Generals = append(rsp.Generals, toGeneral(g))
	}
	return rsp, code.OK
}

func (p *actorPlayer) drawGeneral(req *protocol.DrawGeneralReq) (*protocol.DrawGeneralRsp, int32) {
	cost := req.DrawTimes * 50
	if p.resources.Gold < cost {
		return nil, code.GoldNotEnough
	}
	p.resources.Gold -= cost

	drawn := make([]protocol.General, 0)
	for i := 0; i < req.DrawTimes; i++ {
		g := &model.General{
			Id:            int(time.Now().UnixNano()%1000000) + i,
			RId:           p.rid,
			CfgId:         (i % 10) + 1,
			PhysicalPower: 100,
			Level:         1,
			StarLv:        1,
			Star:          1,
			State:         model.GeneralNormal,
			CreatedAt:     time.Now(),
		}
		_ = p.generalEntry.Insert(g)
		p.generals = append(p.generals, g)
		drawn = append(drawn, toGeneral(g))
	}

	return &protocol.DrawGeneralRsp{Generals: drawn}, code.OK
}

func (p *actorPlayer) composeGeneral(req *protocol.ComposeGeneralReq) (*protocol.ComposeGeneralRsp, int32) {
	return &protocol.ComposeGeneralRsp{Generals: make([]protocol.General, 0)}, code.OK
}

func (p *actorPlayer) addPrGeneral(req *protocol.AddPrGeneralReq) (*protocol.AddPrGeneralRsp, int32) {
	g := p.findGeneral(req.CompId)
	if g == nil {
		return nil, code.GeneralNotFound
	}

	totalAdd := req.ForceAdd + req.StrategyAdd + req.DefenseAdd + req.SpeedAdd + req.DestroyAdd
	if g.HasPrPoint < totalAdd || totalAdd <= 0 {
		return nil, code.InvalidParam
	}

	g.ForceAdded += req.ForceAdd
	g.StrategyAdded += req.StrategyAdd
	g.DefenseAdded += req.DefenseAdd
	g.SpeedAdded += req.SpeedAdd
	g.DestroyAdded += req.DestroyAdd
	g.HasPrPoint -= totalAdd
	g.UsePrPoint += totalAdd
	_ = p.generalEntry.Update(g)

	return &protocol.AddPrGeneralRsp{Generals: toGeneral(g)}, code.OK
}

func (p *actorPlayer) convertGeneral(req *protocol.ConvertReq) (*protocol.ConvertRsp, int32) {
	addGold := 0
	converted := make([]int, 0)
	for _, gid := range req.GIds {
		g := p.findGeneral(gid)
		if g == nil {
			continue
		}
		gold := g.Level * 10
		addGold += gold
		g.State = model.GeneralConverted
		_ = p.generalEntry.Update(g)
		p.removeGeneral(gid)
		converted = append(converted, gid)
	}
	p.resources.Gold += addGold

	return &protocol.ConvertRsp{
		GIds:    converted,
		Gold:    p.resources.Gold,
		AddGold: addGold,
	}, code.OK
}

func (p *actorPlayer) upSkill(req *protocol.UpDownSkillReq) (*protocol.UpDownSkillRsp, int32) {
	return &protocol.UpDownSkillRsp{GId: req.GId, CfgId: req.CfgId, Pos: req.Pos}, code.OK
}

func (p *actorPlayer) downSkill(req *protocol.UpDownSkillReq) (*protocol.UpDownSkillRsp, int32) {
	return &protocol.UpDownSkillRsp{GId: req.GId, CfgId: req.CfgId, Pos: req.Pos}, code.OK
}

func (p *actorPlayer) lvSkill(req *protocol.LvSkillReq) (*protocol.LvSkillRsp, int32) {
	return &protocol.LvSkillRsp{GId: req.GId, Pos: req.Pos}, code.OK
}

// --- Army handlers ---

func (p *actorPlayer) armyList(req *protocol.ArmyListReq) (*protocol.ArmyListRsp, int32) {
	rsp := &protocol.ArmyListRsp{CityId: req.CityId, Armys: make([]protocol.Army, 0)}
	for _, a := range p.armies {
		if a.CityId == req.CityId {
			rsp.Armys = append(rsp.Armys, toArmy(a))
		}
	}
	return rsp, code.OK
}

func (p *actorPlayer) armyOne(req *protocol.ArmyOneReq) (*protocol.ArmyOneRsp, int32) {
	for _, a := range p.armies {
		if a.CityId == req.CityId && a.Order == req.Order {
			return &protocol.ArmyOneRsp{Army: toArmy(a)}, code.OK
		}
	}
	return nil, code.ArmyNotFound
}

func (p *actorPlayer) dispose(req *protocol.DisposeReq) (*protocol.DisposeRsp, int32) {
	var target *model.Army
	for _, a := range p.armies {
		if a.CityId == req.CityId && a.Order == req.Order {
			target = a
			break
		}
	}
	if target == nil {
		return nil, code.ArmyNotFound
	}

	if req.Position < 0 || req.Position > 2 {
		return nil, code.InvalidParam
	}

	if req.GeneralId == 0 {
		target.Generals[req.Position] = 0
		target.Soldiers[req.Position] = 0
	} else {
		g := p.findGeneral(req.GeneralId)
		if g == nil {
			return nil, code.GeneralNotFound
		}
		target.Generals[req.Position] = req.GeneralId
	}

	_ = p.armyEntry.Update(target)
	return &protocol.DisposeRsp{Army: toArmy(target)}, code.OK
}

func (p *actorPlayer) conscript(req *protocol.ConscriptReq) (*protocol.ConscriptRsp, int32) {
	var target *model.Army
	for _, a := range p.armies {
		if a.Id == req.ArmyId {
			target = a
			break
		}
	}
	if target == nil {
		return nil, code.ArmyNotFound
	}

	totalCost := 0
	for i := 0; i < 3 && i < len(req.Cnts); i++ {
		totalCost += req.Cnts[i] * 2
	}

	if p.resources.Gold < totalCost {
		return nil, code.GoldNotEnough
	}

	p.resources.Gold -= totalCost
	now := time.Now().Unix()
	for i := 0; i < 3 && i < len(req.Cnts); i++ {
		if req.Cnts[i] > 0 {
			target.Soldiers[i] += req.Cnts[i]
			target.ConscriptTimes[i] = now + int64(req.Cnts[i])
			target.ConscriptCnts[i] = req.Cnts[i]
		}
	}

	_ = p.armyEntry.Update(target)
	return &protocol.ConscriptRsp{
		Army:    toArmy(target),
		RoleRes: toRoleRes(p.resources),
	}, code.OK
}

func (p *actorPlayer) assignArmy(req *protocol.AssignArmyReq) (*protocol.AssignArmyRsp, int32) {
	var target *model.Army
	for _, a := range p.armies {
		if a.Id == req.ArmyId {
			target = a
			break
		}
	}
	if target == nil {
		return nil, code.ArmyNotFound
	}

	if target.Generals[0] == 0 {
		return nil, code.ArmyNotMain
	}

	target.Cmd = req.Cmd
	target.ToX = req.X
	target.ToY = req.Y

	var fromCity *model.MapRoleCity
	for _, c := range p.cities {
		if c.CityId == target.CityId {
			fromCity = c
			break
		}
	}
	if fromCity != nil {
		target.FromX = fromCity.X
		target.FromY = fromCity.Y
	}

	now := time.Now().Unix()
	target.Start = now
	target.End = now + 30

	_ = p.armyEntry.Update(target)

	warActorPath := cfacade.NewPath(p.App().NodeID(), "war")
	p.Call(warActorPath, "scheduleArmy", target)

	return &protocol.AssignArmyRsp{Army: toArmy(target)}, code.OK
}

// --- Facility handlers ---

func (p *actorPlayer) facilityList(req *protocol.FacilitiesReq) (*protocol.FacilitiesRsp, int32) {
	rsp := &protocol.FacilitiesRsp{CityId: req.CityId, Facilities: make([]protocol.Facility, 0)}
	for _, f := range p.facilities {
		if f.CityId == req.CityId {
			for _, fac := range f.Facilities {
				rsp.Facilities = append(rsp.Facilities, protocol.Facility{
					Name:   fac.Name,
					Level:  fac.PrivateLevel,
					Type:   fac.Type,
					UpTime: fac.UpTime,
				})
			}
		}
	}
	return rsp, code.OK
}

func (p *actorPlayer) upFacility(req *protocol.UpFacilityReq) (*protocol.UpFacilityRsp, int32) {
	return &protocol.UpFacilityRsp{
		CityId:  req.CityId,
		RoleRes: toRoleRes(p.resources),
	}, code.OK
}

// --- Interior handlers ---

func (p *actorPlayer) collect(_ *protocol.CollectionReq) (*protocol.CollectionRsp, int32) {
	if p.attribute == nil {
		p.attribute = &model.RoleAttribute{RId: p.rid}
	}

	maxTimes := int8(5)
	if int8(p.attribute.CollectTimes) >= maxTimes {
		return nil, code.OutCollectTimesLimit
	}

	gold := 100
	p.resources.Gold += gold
	p.attribute.CollectTimes++
	p.attribute.LastCollectTime = time.Now().Unix()

	return &protocol.CollectionRsp{
		Gold:     gold,
		Limit:    maxTimes,
		CurTimes: int8(p.attribute.CollectTimes),
		NextTime: p.attribute.LastCollectTime + 300,
	}, code.OK
}

func (p *actorPlayer) openCollect(_ *protocol.OpenCollectionReq) (*protocol.OpenCollectionRsp, int32) {
	maxTimes := int8(5)
	curTimes := int8(0)
	var nextTime int64
	if p.attribute != nil {
		curTimes = int8(p.attribute.CollectTimes)
		nextTime = p.attribute.LastCollectTime + 300
	}
	return &protocol.OpenCollectionRsp{
		Limit:    maxTimes,
		CurTimes: curTimes,
		NextTime: nextTime,
	}, code.OK
}

func (p *actorPlayer) transform(_ *protocol.TransformReq) (*protocol.TransformRsp, int32) {
	return &protocol.TransformRsp{}, code.OK
}

// --- Skill handlers ---

func (p *actorPlayer) skillList(_ *protocol.SkillListReq) (*protocol.SkillListRsp, int32) {
	rsp := &protocol.SkillListRsp{List: make([]protocol.Skill, 0, len(p.skills))}
	for _, s := range p.skills {
		rsp.List = append(rsp.List, protocol.Skill{
			Id:       s.Id,
			CfgId:    s.CfgId,
			Generals: s.BelongGenerals,
		})
	}
	return rsp, code.OK
}

// --- War report handlers ---

func (p *actorPlayer) warReport(_ *protocol.WarReportReq) (*protocol.WarReportRsp, int32) {
	reports, _ := p.warReportEntry.FindByRId(p.rid, 100)
	rsp := &protocol.WarReportRsp{List: make([]protocol.WarReport, 0)}
	for _, r := range reports {
		rsp.List = append(rsp.List, protocol.WarReport{
			Id:                r.Id,
			AttackRid:         r.AttackRid,
			DefenseRid:        r.DefenseRid,
			BegAttackArmy:     r.BegAttackArmy,
			BegDefenseArmy:    r.BegDefenseArmy,
			EndAttackArmy:     r.EndAttackArmy,
			EndDefenseArmy:    r.EndDefenseArmy,
			BegAttackGeneral:  r.BegAttackGeneral,
			BegDefenseGeneral: r.BegDefenseGeneral,
			EndAttackGeneral:  r.EndAttackGeneral,
			EndDefenseGeneral: r.EndDefenseGeneral,
			Result:            int(r.Result),
			Rounds:            fmt.Sprintf("%d", r.Rounds),
			AttackIsRead:      r.AttackIsRead,
			DefenseIsRead:     r.DefenseIsRead,
			DestroyDurable:    r.DestroyDurable,
			Occupy:            r.Occupy,
			X:                 r.X,
			Y:                 r.Y,
			CTime:             int(r.CreatedAt.Unix()),
		})
	}
	return rsp, code.OK
}

func (p *actorPlayer) warRead(req *protocol.WarReadReq) (*protocol.WarReadRsp, int32) {
	_ = p.warReportEntry.MarkRead(int(req.Id), true)
	return &protocol.WarReadRsp{Id: req.Id}, code.OK
}

// --- Remote handlers ---

func (p *actorPlayer) updateArmyResult(result *model.Army) {
	for i, a := range p.armies {
		if a.Id == result.Id {
			p.armies[i] = result
			break
		}
	}
}

// --- Helpers ---

func (p *actorPlayer) findGeneral(id int) *model.General {
	for _, g := range p.generals {
		if g.Id == id {
			return g
		}
	}
	return nil
}

func (p *actorPlayer) removeGeneral(id int) {
	for i, g := range p.generals {
		if g.Id == id {
			p.generals = append(p.generals[:i], p.generals[i+1:]...)
			return
		}
	}
}
