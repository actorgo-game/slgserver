package controller

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/logic"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorNationMap struct {
	cactor.Base
}

func NewActorNationMap() *ActorNationMap  { return &ActorNationMap{} }
func (p *ActorNationMap) AliasID() string { return "nationMap" }
func (p *ActorNationMap) OnInit()         { clog.Info("[ActorNationMap] initialized") }

func (p *ActorNationMap) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &NationMapChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type NationMapChild struct{ userActor }

func (p *NationMapChild) OnInit() {
	p.Remote().Register("config", p.config)
	p.Remote().Register("scan", p.scan)
	p.Remote().Register("scanBlock", p.scanBlock)
	p.Remote().Register("giveUp", p.giveUp)
	p.Remote().Register("build", p.build)
	p.Remote().Register("upBuild", p.upBuild)
	p.Remote().Register("delBuild", p.delBuild)
}

func (p *NationMapChild) config(_ *protocol.ConfigReq) (*protocol.ConfigRsp, int32) {
	rsp := &protocol.ConfigRsp{}
	m := static_conf.MapBuildConf.Cfg
	rsp.Confs = make([]protocol.Conf, len(m))
	for i, v := range m {
		rsp.Confs[i] = protocol.Conf{
			Type:     v.Type,
			Name:     v.Name,
			Level:    v.Level,
			Defender: v.Defender,
			Durable:  v.Durable,
			Grain:    v.Grain,
			Iron:     v.Iron,
			Stone:    v.Stone,
			Wood:     v.Wood,
		}
	}
	return rsp, code.OK
}

func (p *NationMapChild) scan(req *protocol.ScanReq) (*protocol.ScanRsp, int32) {
	rsp := &protocol.ScanRsp{
		MRBuilds: make([]protocol.MapRoleBuild, 0),
		MCBuilds: make([]protocol.MapRoleCity, 0),
		Armys:    make([]protocol.Army, 0),
	}
	for _, v := range mgr.RBMgr.Scan(req.X, req.Y) {
		rsp.MRBuilds = append(rsp.MRBuilds, v.ToProto().(protocol.MapRoleBuild))
	}
	for _, v := range mgr.RCMgr.Scan(req.X, req.Y) {
		rsp.MCBuilds = append(rsp.MCBuilds, v.ToProto().(protocol.MapRoleCity))
	}
	return rsp, code.OK
}

func (p *NationMapChild) scanBlock(req *protocol.ScanBlockReq) (*protocol.ScanRsp, int32) {
	rsp := &protocol.ScanRsp{
		MRBuilds: make([]protocol.MapRoleBuild, 0),
		MCBuilds: make([]protocol.MapRoleCity, 0),
		Armys:    make([]protocol.Army, 0),
	}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	for _, v := range mgr.RBMgr.ScanBlock(req.X, req.Y, req.Length) {
		rsp.MRBuilds = append(rsp.MRBuilds, v.ToProto().(protocol.MapRoleBuild))
	}
	for _, v := range mgr.RCMgr.ScanBlock(req.X, req.Y, req.Length) {
		rsp.MCBuilds = append(rsp.MCBuilds, v.ToProto().(protocol.MapRoleCity))
	}
	if logic.ArmyLogic != nil {
		for _, v := range logic.ArmyLogic.ScanBlock(role.RId, req.X, req.Y, req.Length) {
			rsp.Armys = append(rsp.Armys, v.ToProto().(protocol.Army))
		}
	}
	return rsp, code.OK
}

func (p *NationMapChild) giveUp(req *protocol.GiveUpReq) (*protocol.GiveUpRsp, int32) {
	rsp := &protocol.GiveUpRsp{X: req.X, Y: req.Y}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RBMgr.BuildIsRId(req.X, req.Y, role.RId) {
		return rsp, code.BuildNotMe
	}
	return rsp, mgr.RBMgr.GiveUp(req.X, req.Y)
}

func (p *NationMapChild) build(req *protocol.BuildReq) (*protocol.BuildRsp, int32) {
	rsp := &protocol.BuildRsp{X: req.X, Y: req.Y, Type: req.Type}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RBMgr.BuildIsRId(req.X, req.Y, role.RId) {
		return rsp, code.BuildNotMe
	}
	b, ok := mgr.RBMgr.PositionBuild(req.X, req.Y)
	if !ok {
		return rsp, code.BuildNotMe
	}
	if !b.IsResBuild() || b.IsBusy() {
		return rsp, code.CanNotBuildNew
	}
	if mgr.RBMgr.RoleFortressCnt(role.RId) >= static_conf.Basic.Build.FortressLimit {
		return rsp, code.CanNotBuildNew
	}
	cfg, ok := static_conf.MapBCConf.BuildConfig(req.Type, 1)
	if !ok {
		return rsp, code.InvalidParam
	}
	if c := mgr.RResMgr.TryUseNeed(role.RId, cfg.Need); c != code.OK {
		return rsp, c
	}
	mbCfg := model.MapBuildCfg{
		Type: cfg.Type, Level: cfg.Level, Name: cfg.Name,
		Durable: cfg.Durable, Defender: cfg.Defender, Time: cfg.Time,
	}
	b.BuildOrUp(mbCfg)
	b.SyncExecute()
	return rsp, code.OK
}

func (p *NationMapChild) upBuild(req *protocol.UpBuildReq) (*protocol.UpBuildRsp, int32) {
	rsp := &protocol.UpBuildRsp{X: req.X, Y: req.Y}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RBMgr.BuildIsRId(req.X, req.Y, role.RId) {
		return rsp, code.BuildNotMe
	}
	b, ok := mgr.RBMgr.PositionBuild(req.X, req.Y)
	if !ok {
		return rsp, code.BuildNotMe
	}
	if !b.IsHaveModifyLVAuth() || b.IsInGiveUp() || b.IsBusy() {
		return rsp, code.CanNotUpBuild
	}
	cfg, ok := static_conf.MapBCConf.BuildConfig(b.Type, b.Level+1)
	if !ok {
		return rsp, code.InvalidParam
	}
	if c := mgr.RResMgr.TryUseNeed(role.RId, cfg.Need); c != code.OK {
		return rsp, c
	}
	mbCfg := model.MapBuildCfg{
		Type: cfg.Type, Level: cfg.Level, Name: cfg.Name,
		Durable: cfg.Durable, Defender: cfg.Defender, Time: cfg.Time,
	}
	b.BuildOrUp(mbCfg)
	b.SyncExecute()
	rsp.Build = b.ToProto().(protocol.MapRoleBuild)
	return rsp, code.OK
}

func (p *NationMapChild) delBuild(req *protocol.UpBuildReq) (*protocol.UpBuildRsp, int32) {
	rsp := &protocol.UpBuildRsp{X: req.X, Y: req.Y}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RBMgr.BuildIsRId(req.X, req.Y, role.RId) {
		return rsp, code.BuildNotMe
	}
	c := mgr.RBMgr.Destroy(req.X, req.Y)
	if b, ok := mgr.RBMgr.PositionBuild(req.X, req.Y); ok {
		rsp.Build = b.ToProto().(protocol.MapRoleBuild)
	}
	return rsp, c
}
