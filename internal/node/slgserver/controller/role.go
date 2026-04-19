package controller

import (
	"math/rand"
	"time"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/global"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/pos"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/protocol"
)

// ActorRole 与原 controller.Role 对应。AliasID="role"。
type ActorRole struct {
	cactor.Base
}

func NewActorRole() *ActorRole       { return &ActorRole{} }
func (p *ActorRole) AliasID() string { return "role" }
func (p *ActorRole) OnInit()         { clog.Info("[ActorRole] initialized") }

func (p *ActorRole) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	child := &RoleChild{}
	a, err := p.Child().Create(m.TargetPath().ChildID, child)
	if err != nil {
		clog.Warn("[ActorRole] OnFindChild err=%v cid=%s", err, m.TargetPath().ChildID)
		return nil, false
	}
	return a, true
}

type RoleChild struct {
	userActor
}

func (p *RoleChild) OnInit() {
	p.Remote().Register("enterServer", p.enterServer)
	p.Remote().Register("create", p.create)
	p.Remote().Register("roleList", p.roleList)
	p.Remote().Register("myCity", p.myCity)
	p.Remote().Register("myRoleRes", p.myRoleRes)
	p.Remote().Register("myRoleBuild", p.myRoleBuild)
	p.Remote().Register("myProperty", p.myProperty)
	p.Remote().Register("upPosition", p.upPosition)
	p.Remote().Register("posTagList", p.posTagList)
	p.Remote().Register("opPosTag", p.opPosTag)
}

func (p *RoleChild) create(req *protocol.CreateRoleReq) (*protocol.CreateRoleRsp, int32) {
	rsp := &protocol.CreateRoleRsp{}
	uid := p.UId()
	if uid == 0 {
		return rsp, code.PlayerIDError
	}

	if exists := mgr.RMgr.ListByUId(uid); len(exists) > 0 {
		return rsp, code.RoleAlreadyCreate
	}

	r := &model.Role{
		UId:       uid,
		HeadId:    req.HeadId,
		Sex:       req.Sex,
		NickName:  req.NickName,
		CreatedAt: time.Now(),
	}
	if err := mgr.RMgr.Insert(r); err != nil {
		clog.Warn("[RoleChild] create err=%v", err)
		return rsp, code.DBError
	}
	rsp.Role = r.ToProto().(protocol.Role)
	return rsp, code.OK
}

func (p *RoleChild) roleList(_ *protocol.RoleListReq) (*protocol.RoleListRsp, int32) {
	rsp := &protocol.RoleListRsp{Roles: make([]protocol.Role, 0)}
	rs := mgr.RMgr.ListByUId(p.UId())
	for _, r := range rs {
		rsp.Roles = append(rsp.Roles, r.ToProto().(protocol.Role))
	}
	return rsp, code.OK
}

func (p *RoleChild) enterServer(req *protocol.EnterServerReq) (*protocol.EnterServerRsp, int32) {
	rsp := &protocol.EnterServerRsp{Time: time.Now().UnixNano() / 1e6}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	rsp.Role = role.ToProto().(protocol.Role)

	roleRes, ok := mgr.RResMgr.Get(role.RId)
	if !ok {
		roleRes = &model.RoleRes{
			RId:    role.RId,
			Wood:   static_conf.Basic.Role.Wood,
			Iron:   static_conf.Basic.Role.Iron,
			Stone:  static_conf.Basic.Role.Stone,
			Grain:  static_conf.Basic.Role.Grain,
			Gold:   static_conf.Basic.Role.Gold,
			Decree: static_conf.Basic.Role.Decree,
		}
		if err := mgr.RResMgr.Insert(roleRes); err != nil {
			clog.Warn("[RoleChild] insert role res err=%v", err)
			return rsp, code.DBError
		}
	}
	rsp.RoleRes = roleRes.ToProto().(protocol.RoleRes)

	if _, ok := mgr.RAttrMgr.TryCreate(role.RId); !ok {
		return rsp, code.DBError
	}

	if _, ok := mgr.RCMgr.GetByRId(role.RId); !ok {
		for {
			x := rand.Intn(global.MapWith)
			y := rand.Intn(global.MapHeight)
			if mgr.NMMgr.IsCanBuildCity(x, y) {
				c := &model.MapRoleCity{
					RId:        role.RId,
					X:          x,
					Y:          y,
					IsMain:     1,
					CurDurable: static_conf.Basic.City.Durable,
					Name:       role.NickName,
					CreatedAt:  time.Now(),
				}
				if err := mgr.RCMgr.Insert(c); err != nil {
					return rsp, code.DBError
				}
				mgr.RFMgr.GetAndTryCreate(c.CityId, c.RId)
				break
			}
		}
	}
	rsp.Token = ""
	_ = req
	return rsp, code.OK
}

func (p *RoleChild) myCity(_ *protocol.MyCityReq) (*protocol.MyCityRsp, int32) {
	rsp := &protocol.MyCityRsp{Citys: make([]protocol.MapRoleCity, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if cs, ok := mgr.RCMgr.GetByRId(role.RId); ok {
		for _, v := range cs {
			rsp.Citys = append(rsp.Citys, v.ToProto().(protocol.MapRoleCity))
		}
	}
	return rsp, code.OK
}

func (p *RoleChild) myRoleRes(_ *protocol.MyRoleResReq) (*protocol.MyRoleResRsp, int32) {
	rsp := &protocol.MyRoleResRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	rr, ok := mgr.RResMgr.Get(role.RId)
	if !ok {
		return rsp, code.RoleNotExist
	}
	rsp.RoleRes = rr.ToProto().(protocol.RoleRes)
	return rsp, code.OK
}

func (p *RoleChild) myProperty(_ *protocol.MyRolePropertyReq) (*protocol.MyRolePropertyRsp, int32) {
	rsp := &protocol.MyRolePropertyRsp{
		Citys:    make([]protocol.MapRoleCity, 0),
		MRBuilds: make([]protocol.MapRoleBuild, 0),
		Generals: make([]protocol.General, 0),
		Armys:    make([]protocol.Army, 0),
	}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if cs, ok := mgr.RCMgr.GetByRId(role.RId); ok {
		for _, v := range cs {
			rsp.Citys = append(rsp.Citys, v.ToProto().(protocol.MapRoleCity))
		}
	}
	if bs, ok := mgr.RBMgr.GetRoleBuild(role.RId); ok {
		for _, v := range bs {
			rsp.MRBuilds = append(rsp.MRBuilds, v.ToProto().(protocol.MapRoleBuild))
		}
	}
	if rr, ok := mgr.RResMgr.Get(role.RId); ok {
		rsp.RoleRes = rr.ToProto().(protocol.RoleRes)
	} else {
		return rsp, code.RoleNotExist
	}
	if gs, ok := mgr.GMgr.GetOrCreateByRId(role.RId); ok {
		for _, v := range gs {
			rsp.Generals = append(rsp.Generals, v.ToProto().(protocol.General))
		}
	} else {
		return rsp, code.DBError
	}
	if as, ok := mgr.AMgr.GetByRId(role.RId); ok {
		for _, v := range as {
			rsp.Armys = append(rsp.Armys, v.ToProto().(protocol.Army))
		}
	}
	return rsp, code.OK
}

func (p *RoleChild) myRoleBuild(_ *protocol.MyRoleBuildReq) (*protocol.MyRoleBuildRsp, int32) {
	rsp := &protocol.MyRoleBuildRsp{MRBuilds: make([]protocol.MapRoleBuild, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if bs, ok := mgr.RBMgr.GetRoleBuild(role.RId); ok {
		for _, v := range bs {
			rsp.MRBuilds = append(rsp.MRBuilds, v.ToProto().(protocol.MapRoleBuild))
		}
	}
	return rsp, code.OK
}

func (p *RoleChild) upPosition(req *protocol.UpPositionReq) (*protocol.UpPositionRsp, int32) {
	rsp := &protocol.UpPositionRsp{X: req.X, Y: req.Y}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	pos.RPMgr.Push(req.X, req.Y, role.RId)
	return rsp, code.OK
}

func (p *RoleChild) posTagList(_ *protocol.PosTagListReq) (*protocol.PosTagListRsp, int32) {
	rsp := &protocol.PosTagListRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	attr, ok := mgr.RAttrMgr.Get(role.RId)
	if !ok {
		return rsp, code.RoleNotExist
	}
	rsp.PosTags = toProtoPosTags(attr.PosTagArray)
	return rsp, code.OK
}

func (p *RoleChild) opPosTag(req *protocol.PosTagReq) (*protocol.PosTagRsp, int32) {
	rsp := &protocol.PosTagRsp{X: req.X, Y: req.Y, Type: req.Type, Name: req.Name}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	attr, ok := mgr.RAttrMgr.Get(role.RId)
	if !ok {
		return rsp, code.RoleNotExist
	}
	switch req.Type {
	case 0:
		attr.RemovePosTag(req.X, req.Y)
		attr.SyncExecute()
	case 1:
		limit := static_conf.Basic.Role.PosTagLimit
		if int(limit) >= len(attr.PosTagArray) {
			attr.AddPosTag(req.X, req.Y, req.Name)
			attr.SyncExecute()
		} else {
			return rsp, code.OutPosTagLimit
		}
	default:
		return rsp, code.InvalidParam
	}
	return rsp, code.OK
}

func toProtoPosTags(src []model.PosTag) []protocol.PosTag {
	out := make([]protocol.PosTag, len(src))
	for i, v := range src {
		out[i] = protocol.PosTag{X: v.X, Y: v.Y, Name: v.Name}
	}
	return out
}
