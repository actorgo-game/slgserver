package controller

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorCity struct {
	cactor.Base
}

func NewActorCity() *ActorCity       { return &ActorCity{} }
func (p *ActorCity) AliasID() string { return "city" }
func (p *ActorCity) OnInit()         { clog.Info("[ActorCity] initialized") }

func (p *ActorCity) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &CityChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type CityChild struct{ userActor }

func (p *CityChild) OnInit() {
	p.Remote().Register("facilities", p.facilities)
	p.Remote().Register("upFacility", p.upFacility)
}

func (p *CityChild) facilities(req *protocol.FacilitiesReq) (*protocol.FacilitiesRsp, int32) {
	rsp := &protocol.FacilitiesRsp{CityId: req.CityId}
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
	f, ok := mgr.RFMgr.Get(req.CityId)
	if !ok {
		return rsp, code.CityNotExist
	}
	t := f.Facility()
	rsp.Facilities = make([]protocol.Facility, len(t))
	for i, v := range t {
		rsp.Facilities[i] = protocol.Facility{
			Name:   v.Name,
			Level:  v.GetLevel(),
			Type:   v.Type,
			UpTime: v.UpTime,
		}
	}
	return rsp, code.OK
}

func (p *CityChild) upFacility(req *protocol.UpFacilityReq) (*protocol.UpFacilityRsp, int32) {
	rsp := &protocol.UpFacilityRsp{CityId: req.CityId}
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
	if _, ok := mgr.RFMgr.Get(req.CityId); !ok {
		return rsp, code.CityNotExist
	}
	out, ec := mgr.RFMgr.UpFacility(role.RId, req.CityId, req.FType)
	if ec != code.OK {
		return rsp, ec
	}
	rsp.Facility = protocol.Facility{
		Name:   out.Name,
		Level:  out.GetLevel(),
		Type:   out.Type,
		UpTime: out.UpTime,
	}
	if rr, ok := mgr.RResMgr.Get(role.RId); ok {
		rsp.RoleRes = rr.ToProto().(protocol.RoleRes)
	}
	return rsp, code.OK
}
