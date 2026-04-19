package controller

import (
	"time"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/global"
	"github.com/llr104/slgserver/internal/node/slgserver/logic"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/check"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/general"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorArmy struct {
	cactor.Base
}

func NewActorArmy() *ActorArmy       { return &ActorArmy{} }
func (p *ActorArmy) AliasID() string { return "army" }
func (p *ActorArmy) OnInit()         { clog.Info("[ActorArmy] initialized") }

func (p *ActorArmy) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &ArmyChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type ArmyChild struct{ userActor }

func (p *ArmyChild) OnInit() {
	p.Remote().Register("myList", p.myList)
	p.Remote().Register("myOne", p.myOne)
	p.Remote().Register("dispose", p.dispose)
	p.Remote().Register("conscript", p.conscript)
	p.Remote().Register("assign", p.assign)
}

func (p *ArmyChild) myList(req *protocol.ArmyListReq) (*protocol.ArmyListRsp, int32) {
	rsp := &protocol.ArmyListRsp{CityId: req.CityId, Armys: make([]protocol.Army, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	city, ok := mgr.RCMgr.Get(req.CityId)
	if !ok {
		return rsp, code.CityNotExist
	}
	if city.RId != role.RId {
		return rsp, code.CityNotMe
	}
	as, _ := mgr.AMgr.GetByCity(req.CityId)
	for _, v := range as {
		rsp.Armys = append(rsp.Armys, v.ToProto().(protocol.Army))
	}
	return rsp, code.OK
}

func (p *ArmyChild) myOne(req *protocol.ArmyOneReq) (*protocol.ArmyOneRsp, int32) {
	rsp := &protocol.ArmyOneRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	city, ok := mgr.RCMgr.Get(req.CityId)
	if !ok {
		return rsp, code.CityNotExist
	}
	if city.RId != role.RId {
		return rsp, code.CityNotMe
	}
	a, ok := mgr.AMgr.GetByCityOrder(req.CityId, req.Order)
	if !ok {
		return rsp, code.ArmyNotFound
	}
	rsp.Army = a.ToProto().(protocol.Army)
	return rsp, code.OK
}

func (p *ArmyChild) dispose(req *protocol.DisposeReq) (*protocol.DisposeRsp, int32) {
	rsp := &protocol.DisposeRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if req.Order <= 0 || req.Order > 5 || req.Position < -1 || req.Position > 2 {
		return rsp, code.InvalidParam
	}
	city, ok := mgr.RCMgr.Get(req.CityId)
	if !ok {
		return rsp, code.CityNotExist
	}
	if city.RId != role.RId {
		return rsp, code.CityNotMe
	}
	jc, ok := mgr.RFMgr.GetFacility(city.CityId, facility.JiaoChang)
	if !ok || jc.GetLevel() < req.Order {
		return rsp, code.ArmyNotEnough
	}
	newG, ok := mgr.GMgr.GetByGId(req.GeneralId)
	if !ok {
		return rsp, code.GeneralNotFound
	}
	if newG.RId != role.RId {
		return rsp, code.GeneralNotMe
	}
	army, err := mgr.AMgr.GetOrCreate(role.RId, req.CityId, req.Order)
	if err != nil {
		return rsp, code.DBError
	}
	if army.FromX != city.X || army.FromY != city.Y {
		return rsp, code.ArmyIsOutside
	}

	if req.Position == -1 {
		for pos, g := range army.Gens {
			if g != nil && g.Id == newG.Id {
				if !army.PositionCanModify(pos) {
					if army.Cmd == model.ArmyCmdConscript {
						return rsp, code.GeneralBusy
					}
					return rsp, code.ArmyBusy
				}
				army.GeneralArray[pos] = 0
				army.SoldierArray[pos] = 0
				army.Gens[pos] = nil
				army.SyncExecute()
				break
			}
		}
		newG.Order = 0
		newG.CityId = 0
		newG.SyncExecute()
	} else {
		if !army.PositionCanModify(req.Position) {
			if army.Cmd == model.ArmyCmdConscript {
				return rsp, code.GeneralBusy
			}
			return rsp, code.ArmyBusy
		}
		if newG.CityId != 0 {
			return rsp, code.GeneralBusy
		}
		if !mgr.AMgr.IsRepeat(role.RId, newG.CfgId) {
			return rsp, code.GeneralRepeat
		}
		tst, ok := mgr.RFMgr.GetFacility(city.CityId, facility.TongShuaiTing)
		if req.Position == 2 && (!ok || tst.GetLevel() < req.Order) {
			return rsp, code.TongShuaiNotEnough
		}
		cost := general.General.Cost(newG.CfgId)
		for i, g := range army.Gens {
			if g == nil || i == req.Position {
				continue
			}
			cost += general.General.Cost(g.CfgId)
		}
		if mgr.GetCityCost(city.CityId) < cost {
			return rsp, code.CostNotEnough
		}
		oldG := army.Gens[req.Position]
		if oldG != nil {
			oldG.CityId = 0
			oldG.Order = 0
			oldG.SyncExecute()
		}
		army.GeneralArray[req.Position] = req.GeneralId
		army.Gens[req.Position] = newG
		army.SoldierArray[req.Position] = 0
		newG.Order = req.Order
		newG.CityId = req.CityId
		newG.SyncExecute()
	}
	army.FromX = city.X
	army.FromY = city.Y
	army.SyncExecute()
	rsp.Army = army.ToProto().(protocol.Army)
	return rsp, code.OK
}

func (p *ArmyChild) conscript(req *protocol.ConscriptReq) (*protocol.ConscriptRsp, int32) {
	rsp := &protocol.ConscriptRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if req.ArmyId <= 0 || len(req.Cnts) != 3 ||
		req.Cnts[0] < 0 || req.Cnts[1] < 0 || req.Cnts[2] < 0 {
		return rsp, code.InvalidParam
	}
	army, ok := mgr.AMgr.Get(req.ArmyId)
	if !ok {
		return rsp, code.ArmyNotFound
	}
	if role.RId != army.RId {
		return rsp, code.ArmyNotMe
	}
	for pos, cnt := range req.Cnts {
		if cnt > 0 {
			if army.Gens[pos] == nil {
				return rsp, code.InvalidParam
			}
			if !army.PositionCanModify(pos) {
				return rsp, code.GeneralBusy
			}
		}
	}
	if mgr.RFMgr.GetFacilityLv(army.CityId, facility.MBS) <= 0 {
		return rsp, code.BuildMBSNotFound
	}
	for i, g := range army.Gens {
		if g == nil {
			continue
		}
		l, e := general.GenBasic.GetLevel(g.Level)
		add := mgr.RFMgr.GetAdditions(army.CityId, facility.TypeSoldierLimit)
		if e != nil {
			return rsp, code.InvalidParam
		}
		if l.Soldiers+add[0] < req.Cnts[i]+army.SoldierArray[i] {
			return rsp, code.OutArmyLimit
		}
	}
	total := 0
	for _, n := range req.Cnts {
		total += n
	}
	cs := static_conf.Basic.ConScript
	nr := facility.NeedRes{
		Grain: total * cs.CostGrain, Wood: total * cs.CostWood,
		Gold: total * cs.CostGold, Iron: total * cs.CostIron,
		Stone: total * cs.CostStone, Decree: 0,
	}
	if c := mgr.RResMgr.TryUseNeed(army.RId, nr); c != code.OK {
		return rsp, code.ResNotEnough
	}
	curTime := time.Now().Unix()
	for i := range army.SoldierArray {
		if req.Cnts[i] > 0 {
			army.ConscriptCntArray[i] = req.Cnts[i]
			army.ConscriptTimeArray[i] = int64(req.Cnts[i]*cs.CostTime) + curTime - 2
		}
	}
	army.Cmd = model.ArmyCmdConscript
	army.SyncExecute()
	rsp.Army = army.ToProto().(protocol.Army)
	if rr, ok := mgr.RResMgr.Get(role.RId); ok {
		rsp.RoleRes = rr.ToProto().(protocol.RoleRes)
	}
	return rsp, code.OK
}

func (p *ArmyChild) assign(req *protocol.AssignArmyReq) (*protocol.AssignArmyRsp, int32) {
	rsp := &protocol.AssignArmyRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	army, ok := mgr.AMgr.Get(req.ArmyId)
	if !ok {
		return rsp, code.ArmyNotFound
	}
	if role.RId != army.RId {
		return rsp, code.ArmyNotMe
	}
	var ec2 int32
	switch req.Cmd {
	case model.ArmyCmdBack:
		ec2 = p.assignBack(army)
	case model.ArmyCmdAttack:
		ec2 = p.assignAttack(req, army, role)
	case model.ArmyCmdDefend:
		ec2 = p.assignDefend(req, army, role)
	case model.ArmyCmdReclamation:
		ec2 = p.assignReclamation(req, army, role)
	case model.ArmyCmdTransfer:
		ec2 = p.assignTransfer(req, army, role)
	default:
		ec2 = code.InvalidParam
	}
	rsp.Army = army.ToProto().(protocol.Army)
	return rsp, ec2
}

func (p *ArmyChild) assignPre(req *protocol.AssignArmyReq, army *model.Army, role *model.Role) int32 {
	if req.X < 0 || req.X >= global.MapWith || req.Y < 0 || req.Y >= global.MapHeight {
		return code.InvalidParam
	}
	if !army.IsCanOutWar() {
		if army.Cmd != model.ArmyCmdIdle {
			return code.ArmyBusy
		}
		return code.ArmyNotMain
	}
	if !army.IsIdle() {
		return code.ArmyBusy
	}
	cfg, ok := mgr.NMMgr.PositionBuild(req.X, req.Y)
	if !ok || cfg.Type == 0 {
		return code.InvalidParam
	}
	if !check.IsCanArrive(req.X, req.Y, role.RId) {
		return code.UnReachable
	}
	return code.OK
}

func (p *ArmyChild) assignAfter(req *protocol.AssignArmyReq, army *model.Army) int32 {
	cost := static_conf.Basic.General.CostPhysicalPower
	if !mgr.GMgr.PhysicalPowerIsEnough(army, cost) {
		return code.PhysicalPowerNotEnough
	}
	if req.Cmd == model.ArmyCmdReclamation || req.Cmd == model.ArmyCmdTransfer {
		dec := static_conf.Basic.General.ReclamationCost
		if !mgr.RResMgr.DecreeIsEnough(army.RId, dec) {
			return code.DecreeNotEnough
		}
		mgr.RResMgr.TryUseDecree(army.RId, dec)
	}
	mgr.GMgr.TryUsePhysicalPower(army, cost)
	army.ToX = req.X
	army.ToY = req.Y
	army.Cmd = req.Cmd
	army.State = model.ArmyRunning
	if global.IsDev() {
		army.Start = time.Now()
		army.End = time.Now().Add(40 * time.Second)
	} else {
		speed := mgr.AMgr.GetSpeed(army)
		t := mgr.TravelTime(speed, army.FromX, army.FromY, army.ToX, army.ToY)
		army.Start = time.Now()
		army.End = time.Now().Add(time.Duration(t) * time.Millisecond)
	}
	if logic.ArmyLogic != nil {
		logic.ArmyLogic.PushAction(army)
	}
	return code.OK
}

func (p *ArmyChild) assignBack(army *model.Army) int32 {
	if logic.ArmyLogic == nil {
		return code.OK
	}
	if army.Cmd == model.ArmyCmdAttack || army.Cmd == model.ArmyCmdDefend ||
		army.Cmd == model.ArmyCmdReclamation {
		logic.ArmyLogic.ArmyBack(army)
	} else if army.IsIdle() {
		if city, ok := mgr.RCMgr.Get(army.CityId); ok {
			if city.X != army.FromX || city.Y != army.FromY {
				logic.ArmyLogic.ArmyBack(army)
			}
		}
	}
	return code.OK
}

func (p *ArmyChild) assignAttack(req *protocol.AssignArmyReq, army *model.Army, role *model.Role) int32 {
	if c := p.assignPre(req, army, role); c != code.OK {
		return c
	}
	if !check.IsCanArrive(req.X, req.Y, role.RId) {
		return code.UnReachable
	}
	if check.IsWarFree(req.X, req.Y) {
		return code.BuildWarFree
	}
	if check.IsCanDefend(req.X, req.Y, role.RId) {
		return code.BuildCanNotAttack
	}
	return p.assignAfter(req, army)
}

func (p *ArmyChild) assignDefend(req *protocol.AssignArmyReq, army *model.Army, role *model.Role) int32 {
	if c := p.assignPre(req, army, role); c != code.OK {
		return c
	}
	if !check.IsCanDefend(req.X, req.Y, role.RId) {
		return code.BuildCanNotDefend
	}
	return p.assignAfter(req, army)
}

func (p *ArmyChild) assignReclamation(req *protocol.AssignArmyReq, army *model.Army, role *model.Role) int32 {
	if c := p.assignPre(req, army, role); c != code.OK {
		return c
	}
	if !mgr.RBMgr.BuildIsRId(req.X, req.Y, role.RId) {
		return code.BuildNotMe
	}
	return p.assignAfter(req, army)
}

func (p *ArmyChild) assignTransfer(req *protocol.AssignArmyReq, army *model.Army, role *model.Role) int32 {
	if c := p.assignPre(req, army, role); c != code.OK {
		return c
	}
	if army.FromX == req.X && army.FromY == req.Y {
		return code.CanNotTransfer
	}
	if !mgr.RBMgr.BuildIsRId(req.X, req.Y, role.RId) {
		return code.BuildNotMe
	}
	b, ok := mgr.RBMgr.PositionBuild(req.X, req.Y)
	if !ok {
		return code.BuildNotMe
	}
	if b.Level <= 0 || !b.IsHasTransferAuth() {
		return code.CanNotTransfer
	}
	cnt := 5
	if b.IsRoleFortress() {
		cnt = static_conf.MapBCConf.GetHoldArmyCnt(b.Type, b.Level)
	}
	if cnt <= mgr.AMgr.BelongPosArmyCnt(b.RId, b.X, b.Y) {
		return code.HoldIsFull
	}
	return p.assignAfter(req, army)
}
