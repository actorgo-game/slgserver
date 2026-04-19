package controller

import (
	"time"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorInterior struct {
	cactor.Base
}

func NewActorInterior() *ActorInterior   { return &ActorInterior{} }
func (p *ActorInterior) AliasID() string { return "interior" }
func (p *ActorInterior) OnInit()         { clog.Info("[ActorInterior] initialized") }

func (p *ActorInterior) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &InteriorChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type InteriorChild struct{ userActor }

func (p *InteriorChild) OnInit() {
	p.Remote().Register("collect", p.collect)
	p.Remote().Register("openCollect", p.openCollect)
	p.Remote().Register("transform", p.transform)
}

func nextCollectTime(last time.Time, interval int, cur, limit int8) int64 {
	if cur >= limit {
		y, m, d := last.Add(24 * time.Hour).Date()
		nt := time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("IST", 3600))
		return nt.UnixNano() / 1e6
	}
	nt := last.Add(time.Duration(interval) * time.Second)
	return nt.UnixNano() / 1e6
}

func (p *InteriorChild) collect(_ *protocol.CollectionReq) (*protocol.CollectionRsp, int32) {
	rsp := &protocol.CollectionRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	roleRes, ok := mgr.RResMgr.Get(role.RId)
	if !ok {
		return rsp, code.DBError
	}
	roleAttr, ok := mgr.RAttrMgr.Get(role.RId)
	if !ok {
		return rsp, code.DBError
	}
	curTime := time.Now()
	last := roleAttr.LastCollectTime
	if curTime.YearDay() != last.YearDay() || curTime.Year() != last.Year() {
		roleAttr.CollectTimes = 0
		roleAttr.LastCollectTime = time.Time{}
	}
	timeLimit := static_conf.Basic.Role.CollectTimesLimit
	if roleAttr.CollectTimes >= timeLimit {
		return rsp, code.OutCollectTimesLimit
	}
	need := last.Add(time.Duration(static_conf.Basic.Role.CollectTimesLimit) * time.Second)
	if curTime.Before(need) {
		return rsp, code.InCdCanNotOperate
	}
	gold := mgr.GetYield(roleRes.RId).Gold
	rsp.Gold = gold
	roleRes.Gold += gold
	roleRes.SyncExecute()
	roleAttr.LastCollectTime = curTime
	roleAttr.CollectTimes++
	roleAttr.SyncExecute()
	rsp.NextTime = nextCollectTime(roleAttr.LastCollectTime,
		static_conf.Basic.Role.CollectInterval, roleAttr.CollectTimes, timeLimit)
	rsp.CurTimes = roleAttr.CollectTimes
	rsp.Limit = timeLimit
	return rsp, code.OK
}

func (p *InteriorChild) openCollect(_ *protocol.OpenCollectionReq) (*protocol.OpenCollectionRsp, int32) {
	rsp := &protocol.OpenCollectionRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	roleAttr, ok := mgr.RAttrMgr.Get(role.RId)
	if !ok {
		return rsp, code.DBError
	}
	timeLimit := static_conf.Basic.Role.CollectTimesLimit
	rsp.Limit = timeLimit
	rsp.CurTimes = roleAttr.CollectTimes
	if roleAttr.LastCollectTime.IsZero() {
		rsp.NextTime = 0
	} else {
		rsp.NextTime = nextCollectTime(roleAttr.LastCollectTime,
			static_conf.Basic.Role.CollectInterval, roleAttr.CollectTimes, timeLimit)
	}
	return rsp, code.OK
}

func (p *InteriorChild) transform(req *protocol.TransformReq) (*protocol.TransformRsp, int32) {
	rsp := &protocol.TransformRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	roleRes, ok := mgr.RResMgr.Get(role.RId)
	if !ok {
		return rsp, code.DBError
	}
	main, ok := mgr.RCMgr.GetMainCity(role.RId)
	if !ok {
		return rsp, code.DBError
	}
	if mgr.RFMgr.GetFacilityLv(main.CityId, facility.JiShi) <= 0 {
		return rsp, code.NotHasJiShi
	}
	if len(req.From) < 4 || len(req.To) < 4 {
		return rsp, code.InvalidParam
	}
	ret := make([]int, 4)
	for i := 0; i < 4; i++ {
		if req.From[i] > 0 {
			ret[i] = -req.From[i]
		}
		if req.To[i] > 0 {
			ret[i] = req.To[i]
		}
	}
	if roleRes.Wood+ret[0] < 0 || roleRes.Iron+ret[1] < 0 ||
		roleRes.Stone+ret[2] < 0 || roleRes.Grain+ret[3] < 0 {
		return rsp, code.InvalidParam
	}
	roleRes.Wood += ret[0]
	roleRes.Iron += ret[1]
	roleRes.Stone += ret[2]
	roleRes.Grain += ret[3]
	roleRes.SyncExecute()
	return rsp, code.OK
}
