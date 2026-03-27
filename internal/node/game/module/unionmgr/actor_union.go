package unionmgr

import (
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/component"
	"github.com/llr104/slgserver/internal/data/entry"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/protocol"
)

// Internal request types for Remote handlers that need caller identification.
// The player actor includes its RId when calling these handlers.

type unionCreateReq struct {
	RId  int    `json:"rid"`
	Name string `json:"name"`
}

type unionJoinReq struct {
	RId int `json:"rid"`
	Id  int `json:"id"`
}

type unionRIdReq struct {
	RId int `json:"rid"`
}

type unionModNoticeReq struct {
	RId  int    `json:"rid"`
	Text string `json:"text"`
}

type unionKickReq struct {
	CallerRId int `json:"callerRid"`
	RId       int `json:"rid"`
}

type unionAbdicateReq struct {
	CallerRId int `json:"callerRid"`
	RId       int `json:"rid"`
}

type ActorUnion struct {
	cactor.Base

	unions         map[int]*model.Coalition
	coalitionEntry *entry.CoalitionEntry
	applyEntry     *entry.CoalitionApplyEntry
	logEntry       *entry.CoalitionLogEntry
	nextId         int
	serverId       int
}

func NewActorUnion() *ActorUnion {
	return &ActorUnion{
		unions: make(map[int]*model.Coalition),
	}
}

func (p *ActorUnion) AliasID() string {
	return "union"
}

func (p *ActorUnion) OnInit() {
	p.initDB()
	p.loadData()

	p.Remote().Register("create", p.create)
	p.Remote().Register("list", p.list)
	p.Remote().Register("join", p.join)
	p.Remote().Register("verify", p.verify)
	p.Remote().Register("member", p.memberList)
	p.Remote().Register("applyList", p.applyList)
	p.Remote().Register("exit", p.exit)
	p.Remote().Register("dismiss", p.dismiss)
	p.Remote().Register("notice", p.notice)
	p.Remote().Register("modNotice", p.modNotice)
	p.Remote().Register("kick", p.kick)
	p.Remote().Register("appoint", p.appoint)
	p.Remote().Register("abdicate", p.abdicate)
	p.Remote().Register("info", p.info)
	p.Remote().Register("log", p.unionLog)
	p.Remote().Register("getRIdUnion", p.getRIdUnion)

	clog.Info("[ActorUnion] loaded %d coalitions", len(p.unions))
}

func (p *ActorUnion) initDB() {
	mongoComp := p.App().Find(component.MongoComponentName)
	if mongoComp == nil {
		return
	}
	mc := mongoComp.(*component.MongoComponent)
	db := mc.GetDb("slg_db")
	if db == nil {
		return
	}
	p.serverId = p.App().Settings().GetInt("server_id", 1)
	p.coalitionEntry = entry.NewCoalitionEntry(db.Collection("coalitions"), p.serverId)
	p.applyEntry = entry.NewCoalitionApplyEntry(db.Collection("coalition_applies"), p.serverId)
	p.logEntry = entry.NewCoalitionLogEntry(db.Collection("coalition_logs"), p.serverId)
}

func (p *ActorUnion) loadData() {
	if p.coalitionEntry == nil {
		return
	}
	coalitions, err := p.coalitionEntry.FindAllRunning()
	if err != nil {
		clog.Warn("[ActorUnion] load fail: %v", err)
		return
	}
	for _, c := range coalitions {
		p.unions[c.Id] = c
		if c.Id >= p.nextId {
			p.nextId = c.Id + 1
		}
	}
}

func (p *ActorUnion) findUnionByRId(rid int) *model.Coalition {
	for _, u := range p.unions {
		for _, m := range u.Members {
			if m == rid {
				return u
			}
		}
	}
	return nil
}

// --- handlers ---

func (p *ActorUnion) create(req *unionCreateReq) (*protocol.CreateRsp, int32) {
	rid := req.RId
	if rid == 0 {
		return nil, code.RoleNotExist
	}

	if p.findUnionByRId(rid) != nil {
		return nil, code.UnionAlreadyHas
	}

	p.nextId++
	u := &model.Coalition{
		Id:        p.nextId,
		Name:      req.Name,
		Members:   []int{rid},
		CreateId:  rid,
		Chairman:  rid,
		State:     model.UnionRunning,
		CreatedAt: time.Now(),
	}
	p.unions[u.Id] = u
	if p.coalitionEntry != nil {
		_ = p.coalitionEntry.Insert(u)
	}

	return &protocol.CreateRsp{Id: u.Id, Name: u.Name}, code.OK
}

func (p *ActorUnion) list(_ *protocol.ListReq) (*protocol.ListRsp, int32) {
	rsp := &protocol.ListRsp{List: make([]protocol.Union, 0, len(p.unions))}
	for _, u := range p.unions {
		rsp.List = append(rsp.List, toUnionProto(u))
	}
	return rsp, code.OK
}

func (p *ActorUnion) join(req *unionJoinReq) (*protocol.JoinRsp, int32) {
	rid := req.RId
	u, ok := p.unions[req.Id]
	if !ok {
		return nil, code.UnionNotFound
	}

	if p.findUnionByRId(rid) != nil {
		return nil, code.UnionAlreadyHas
	}

	apply := &model.CoalitionApply{
		Id:        int(time.Now().UnixNano() % 1000000),
		UnionId:   u.Id,
		RId:       rid,
		State:     0,
		CreatedAt: time.Now(),
	}
	if p.applyEntry != nil {
		_ = p.applyEntry.Insert(apply)
	}

	return &protocol.JoinRsp{}, code.OK
}

func (p *ActorUnion) verify(req *protocol.VerifyReq) (*protocol.VerifyRsp, int32) {
	return &protocol.VerifyRsp{Id: req.Id, Decide: req.Decide}, code.OK
}

func (p *ActorUnion) memberList(req *protocol.MemberReq) (*protocol.MemberRsp, int32) {
	u, ok := p.unions[req.Id]
	if !ok {
		return nil, code.UnionNotFound
	}

	rsp := &protocol.MemberRsp{Id: u.Id, Members: make([]protocol.Member, 0)}
	for _, rid := range u.Members {
		rsp.Members = append(rsp.Members, protocol.Member{RId: rid})
	}
	return rsp, code.OK
}

func (p *ActorUnion) applyList(req *protocol.ApplyReq) (*protocol.ApplyRsp, int32) {
	rsp := &protocol.ApplyRsp{Id: req.Id, Applys: make([]protocol.ApplyItem, 0)}
	if p.applyEntry != nil {
		applies, _ := p.applyEntry.FindByUnionId(req.Id)
		for _, a := range applies {
			rsp.Applys = append(rsp.Applys, protocol.ApplyItem{Id: a.Id, RId: a.RId})
		}
	}
	return rsp, code.OK
}

func (p *ActorUnion) exit(req *unionRIdReq) (*protocol.ExitRsp, int32) {
	rid := req.RId
	u := p.findUnionByRId(rid)
	if u == nil {
		return nil, code.UnionNotFound
	}
	if u.Chairman == rid {
		return nil, code.UnionNotAllowExit
	}
	for i, m := range u.Members {
		if m == rid {
			u.Members = append(u.Members[:i], u.Members[i+1:]...)
			break
		}
	}
	if p.coalitionEntry != nil {
		_ = p.coalitionEntry.Update(u)
	}
	return &protocol.ExitRsp{}, code.OK
}

func (p *ActorUnion) dismiss(req *unionRIdReq) (*protocol.DismissRsp, int32) {
	rid := req.RId
	u := p.findUnionByRId(rid)
	if u == nil {
		return nil, code.UnionNotFound
	}
	if u.Chairman != rid {
		return nil, code.PermissionDenied
	}
	u.State = model.UnionDismiss
	u.Members = nil
	if p.coalitionEntry != nil {
		_ = p.coalitionEntry.Update(u)
	}
	delete(p.unions, u.Id)
	return &protocol.DismissRsp{}, code.OK
}

func (p *ActorUnion) notice(req *protocol.NoticeReq) (*protocol.NoticeRsp, int32) {
	u, ok := p.unions[req.Id]
	if !ok {
		return nil, code.UnionNotFound
	}
	return &protocol.NoticeRsp{Text: u.Notice}, code.OK
}

func (p *ActorUnion) modNotice(req *unionModNoticeReq) (*protocol.ModNoticeRsp, int32) {
	rid := req.RId
	u := p.findUnionByRId(rid)
	if u == nil {
		return nil, code.UnionNotFound
	}
	if u.Chairman != rid && u.ViceChairman != rid {
		return nil, code.PermissionDenied
	}
	u.Notice = req.Text
	if p.coalitionEntry != nil {
		_ = p.coalitionEntry.Update(u)
	}
	return &protocol.ModNoticeRsp{Id: u.Id, Text: u.Notice}, code.OK
}

func (p *ActorUnion) kick(req *unionKickReq) (*protocol.KickRsp, int32) {
	rid := req.CallerRId
	u := p.findUnionByRId(rid)
	if u == nil || (u.Chairman != rid && u.ViceChairman != rid) {
		return nil, code.PermissionDenied
	}
	for i, m := range u.Members {
		if m == req.RId {
			u.Members = append(u.Members[:i], u.Members[i+1:]...)
			break
		}
	}
	if p.coalitionEntry != nil {
		_ = p.coalitionEntry.Update(u)
	}
	return &protocol.KickRsp{RId: req.RId}, code.OK
}

func (p *ActorUnion) appoint(req *protocol.AppointReq) (*protocol.AppointRsp, int32) {
	return &protocol.AppointRsp{RId: req.RId, Title: req.Title}, code.OK
}

func (p *ActorUnion) abdicate(req *unionAbdicateReq) (*protocol.AbdicateRsp, int32) {
	rid := req.CallerRId
	u := p.findUnionByRId(rid)
	if u == nil || u.Chairman != rid {
		return nil, code.PermissionDenied
	}
	u.Chairman = req.RId
	if p.coalitionEntry != nil {
		_ = p.coalitionEntry.Update(u)
	}
	return &protocol.AbdicateRsp{}, code.OK
}

func (p *ActorUnion) info(req *protocol.InfoReq) (*protocol.InfoRsp, int32) {
	u, ok := p.unions[req.Id]
	if !ok {
		return nil, code.UnionNotFound
	}
	return &protocol.InfoRsp{Id: u.Id, Info: toUnionProto(u)}, code.OK
}

func (p *ActorUnion) unionLog(req *protocol.LogReq) (*protocol.LogRsp, int32) {
	rsp := &protocol.LogRsp{Logs: make([]protocol.UnionLog, 0)}
	if p.logEntry != nil {
		rid := 0
		u := p.findUnionByRId(rid)
		if u != nil {
			logs, _ := p.logEntry.FindByUnionId(u.Id)
			for _, l := range logs {
				rsp.Logs = append(rsp.Logs, protocol.UnionLog{
					OPRId: l.OPRId, TargetId: l.TargetId, State: l.State,
					Des: l.Des, Ctime: l.CreatedAt.Unix(),
				})
			}
		}
	}
	return rsp, code.OK
}

type getRIdUnionReq struct {
	RId int `json:"rid"`
}

type getRIdUnionRsp struct {
	UnionId int `json:"unionId"`
}

func (p *ActorUnion) getRIdUnion(req *getRIdUnionReq) (*getRIdUnionRsp, int32) {
	u := p.findUnionByRId(req.RId)
	if u == nil {
		return &getRIdUnionRsp{}, code.OK
	}
	return &getRIdUnionRsp{UnionId: u.Id}, code.OK
}

func toUnionProto(u *model.Coalition) protocol.Union {
	majors := make([]protocol.Major, 0)
	if u.Chairman != 0 {
		majors = append(majors, protocol.Major{RId: u.Chairman, Title: protocol.UnionChairman})
	}
	if u.ViceChairman != 0 {
		majors = append(majors, protocol.Major{RId: u.ViceChairman, Title: protocol.UnionViceChairman})
	}
	return protocol.Union{
		Id:     u.Id,
		Name:   u.Name,
		Cnt:    len(u.Members),
		Notice: u.Notice,
		Major:  majors,
	}
}
