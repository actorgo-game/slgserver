package controller

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/skill"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorGeneral struct {
	cactor.Base
}

func NewActorGeneral() *ActorGeneral    { return &ActorGeneral{} }
func (p *ActorGeneral) AliasID() string { return "general" }
func (p *ActorGeneral) OnInit()         { clog.Info("[ActorGeneral] initialized") }

func (p *ActorGeneral) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &GeneralChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type GeneralChild struct{ userActor }

func (p *GeneralChild) OnInit() {
	p.Remote().Register("myGenerals", p.myGenerals)
	p.Remote().Register("drawGeneral", p.drawGenerals)
	p.Remote().Register("composeGeneral", p.composeGeneral)
	p.Remote().Register("addPrGeneral", p.addPrGeneral)
	p.Remote().Register("convert", p.convert)
	p.Remote().Register("upSkill", p.upSkill)
	p.Remote().Register("downSkill", p.downSkill)
	p.Remote().Register("lvSkill", p.lvSkill)
}

func (p *GeneralChild) myGenerals(_ *protocol.MyGeneralReq) (*protocol.MyGeneralRsp, int32) {
	rsp := &protocol.MyGeneralRsp{Generals: make([]protocol.General, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	gs, ok := mgr.GMgr.GetOrCreateByRId(role.RId)
	if !ok {
		return rsp, code.DBError
	}
	for _, v := range gs {
		rsp.Generals = append(rsp.Generals, v.ToProto().(protocol.General))
	}
	return rsp, code.OK
}

func (p *GeneralChild) drawGenerals(req *protocol.DrawGeneralReq) (*protocol.DrawGeneralRsp, int32) {
	rsp := &protocol.DrawGeneralRsp{Generals: make([]protocol.General, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	cost := static_conf.Basic.General.DrawGeneralCost * req.DrawTimes
	if !mgr.RResMgr.GoldIsEnough(role.RId, cost) {
		return rsp, code.GoldNotEnough
	}
	limit := static_conf.Basic.General.Limit
	if mgr.GMgr.Count(role.RId)+req.DrawTimes > limit {
		return rsp, code.OutGeneralLimit
	}
	gs, ok := mgr.GMgr.RandCreateGeneral(role.RId, req.DrawTimes)
	if !ok {
		return rsp, code.DBError
	}
	mgr.RResMgr.TryUseGold(role.RId, cost)
	for _, v := range gs {
		rsp.Generals = append(rsp.Generals, v.ToProto().(protocol.General))
	}
	return rsp, code.OK
}

func (p *GeneralChild) composeGeneral(req *protocol.ComposeGeneralReq) (*protocol.ComposeGeneralRsp, int32) {
	rsp := &protocol.ComposeGeneralRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	gs, ok := mgr.GMgr.HasGeneral(role.RId, req.CompId)
	if !ok {
		return rsp, code.GeneralNoHas
	}
	gss, ok := mgr.GMgr.HasGenerals(role.RId, req.GIds)
	if !ok {
		return rsp, code.GeneralNoHas
	}
	for _, v := range gss {
		if v.CfgId != gs.CfgId {
			return rsp, code.GeneralNoSame
		}
	}
	if int(gs.Star-gs.StarLv) < len(gss) {
		return rsp, code.GeneralStarMax
	}
	gs.StarLv += int8(len(gss))
	gs.HasPrPoint += static_conf.Basic.General.PrPoint * len(gss)
	gs.SyncExecute()
	for _, v := range gss {
		v.ParentId = gs.Id
		v.State = model.GeneralComposeStar
		v.SyncExecute()
	}
	rsp.Generals = make([]protocol.General, 0, len(gss)+1)
	for _, v := range gss {
		rsp.Generals = append(rsp.Generals, v.ToProto().(protocol.General))
	}
	rsp.Generals = append(rsp.Generals, gs.ToProto().(protocol.General))
	return rsp, code.OK
}

func (p *GeneralChild) addPrGeneral(req *protocol.AddPrGeneralReq) (*protocol.AddPrGeneralRsp, int32) {
	rsp := &protocol.AddPrGeneralRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	gs, ok := mgr.GMgr.HasGeneral(role.RId, req.CompId)
	if !ok {
		return rsp, code.GeneralNoHas
	}
	all := req.ForceAdd + req.StrategyAdd + req.DefenseAdd + req.SpeedAdd + req.DestroyAdd
	if gs.HasPrPoint < all {
		return rsp, code.DBError
	}
	gs.ForceAdded = req.ForceAdd
	gs.StrategyAdded = req.StrategyAdd
	gs.DefenseAdded = req.DefenseAdd
	gs.SpeedAdded = req.SpeedAdd
	gs.DestroyAdded = req.DestroyAdd
	gs.UsePrPoint = all
	gs.SyncExecute()
	rsp.Generals = gs.ToProto().(protocol.General)
	return rsp, code.OK
}

func (p *GeneralChild) convert(req *protocol.ConvertReq) (*protocol.ConvertRsp, int32) {
	rsp := &protocol.ConvertRsp{GIds: make([]int, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	rr, ok := mgr.RResMgr.Get(role.RId)
	if !ok {
		return rsp, code.DBError
	}
	gold := 0
	for _, gid := range req.GIds {
		g, ok := mgr.GMgr.GetByGId(gid)
		if ok && g.Order == 0 {
			rsp.GIds = append(rsp.GIds, gid)
			gold += 10 * int(g.Star) * (1 + int(g.StarLv))
			g.State = model.GeneralConvert
			g.SyncExecute()
		}
	}
	rr.Gold += gold
	rsp.AddGold = gold
	rsp.Gold = rr.Gold
	rr.SyncExecute()
	return rsp, code.OK
}

func (p *GeneralChild) upSkill(req *protocol.UpDownSkillReq) (*protocol.UpDownSkillRsp, int32) {
	rsp := &protocol.UpDownSkillRsp{Pos: req.Pos, CfgId: req.CfgId, GId: req.GId}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if req.Pos < 0 || req.Pos >= model.SkillLimit {
		return rsp, code.InvalidParam
	}
	g, ok := mgr.GMgr.GetByGId(req.GId)
	if !ok {
		return rsp, code.GeneralNotFound
	}
	if g.RId != role.RId {
		return rsp, code.GeneralNotMe
	}
	sk, ok := mgr.SkillMgr.GetSkillOrCreate(role.RId, req.CfgId)
	if !ok {
		return rsp, code.DBError
	}
	if !sk.IsInLimit() {
		return rsp, code.OutSkillLimit
	}
	if !sk.ArmyIsIn(g.CurArms) {
		return rsp, code.OutArmNotMatch
	}
	if !g.UpSkill(sk.Id, req.CfgId, req.Pos) {
		return rsp, code.UpSkillError
	}
	sk.UpSkill(g.Id)
	g.SyncExecute()
	sk.SyncExecute()
	return rsp, code.OK
}

func (p *GeneralChild) downSkill(req *protocol.UpDownSkillReq) (*protocol.UpDownSkillRsp, int32) {
	rsp := &protocol.UpDownSkillRsp{Pos: req.Pos, CfgId: req.CfgId, GId: req.GId}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if req.Pos < 0 || req.Pos >= model.SkillLimit {
		return rsp, code.InvalidParam
	}
	g, ok := mgr.GMgr.GetByGId(req.GId)
	if !ok {
		return rsp, code.GeneralNotFound
	}
	if g.RId != role.RId {
		return rsp, code.GeneralNotMe
	}
	sk, ok := mgr.SkillMgr.GetSkillOrCreate(role.RId, req.CfgId)
	if !ok {
		return rsp, code.DBError
	}
	if !g.DownSkill(sk.Id, req.Pos) {
		return rsp, code.DownSkillError
	}
	sk.DownSkill(g.Id)
	g.SyncExecute()
	sk.SyncExecute()
	return rsp, code.OK
}

func (p *GeneralChild) lvSkill(req *protocol.LvSkillReq) (*protocol.LvSkillRsp, int32) {
	rsp := &protocol.LvSkillRsp{Pos: req.Pos, GId: req.GId}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	g, ok := mgr.GMgr.GetByGId(req.GId)
	if !ok {
		return rsp, code.GeneralNotFound
	}
	if g.RId != role.RId {
		return rsp, code.GeneralNotMe
	}
	gSkill, err := g.PosSkill(req.Pos)
	if err != nil {
		return rsp, code.PosNotSkill
	}
	skillCfg, ok := skill.Skill.GetCfg(gSkill.CfgId)
	if !ok {
		return rsp, code.PosNotSkill
	}
	if gSkill.Lv > len(skillCfg.Levels) {
		return rsp, code.SkillLevelFull
	}
	gSkill.Lv += 1
	g.SyncExecute()
	return rsp, code.OK
}
