package controller

import (
	"context"
	"time"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/db"
	"github.com/llr104/slgserver/internal/node/slgserver/logic"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorUnion struct {
	cactor.Base
}

func NewActorUnion() *ActorUnion      { return &ActorUnion{} }
func (p *ActorUnion) AliasID() string { return "union" }
func (p *ActorUnion) OnInit()         { clog.Info("[ActorUnion] initialized") }

func (p *ActorUnion) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &UnionChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type UnionChild struct{ userActor }

func (p *UnionChild) OnInit() {
	p.Remote().Register("create", p.create)
	p.Remote().Register("list", p.list)
	p.Remote().Register("join", p.join)
	p.Remote().Register("verify", p.verify)
	p.Remote().Register("member", p.member)
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
}

func (p *UnionChild) majorOf(u *model.Coalition) []protocol.Major {
	out := make([]protocol.Major, 0, 2)
	if r, ok := mgr.RMgr.Get(u.Chairman); ok {
		out = append(out, protocol.Major{Name: r.NickName, RId: r.RId, Title: protocol.UnionChairman})
	}
	if r, ok := mgr.RMgr.Get(u.ViceChairman); ok {
		out = append(out, protocol.Major{Name: r.NickName, RId: r.RId, Title: protocol.UnionViceChairman})
	}
	return out
}

func (p *UnionChild) create(req *protocol.CreateReq) (*protocol.CreateRsp, int32) {
	rsp := &protocol.CreateRsp{Name: req.Name}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionAlreadyHas
	}
	c, ok := mgr.UnionMgr.Create(req.Name, role.RId)
	if !ok {
		return rsp, code.UnionCreateError
	}
	rsp.Id = c.Id
	if logic.Union != nil {
		logic.Union.MemberEnter(role.RId, c.Id)
	}
	model.NewCreate(role.NickName, c.Id, role.RId)
	return rsp, code.OK
}

func (p *UnionChild) list(_ *protocol.ListReq) (*protocol.ListRsp, int32) {
	rsp := &protocol.ListRsp{List: make([]protocol.Union, 0)}
	for _, u := range mgr.UnionMgr.List() {
		pu := u.ToProto().(protocol.Union)
		pu.Major = p.majorOf(u)
		rsp.List = append(rsp.List, pu)
	}
	return rsp, code.OK
}

func (p *UnionChild) join(req *protocol.JoinReq) (*protocol.JoinRsp, int32) {
	rsp := &protocol.JoinRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionAlreadyHas
	}
	u, ok := mgr.UnionMgr.Get(req.Id)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if len(u.MemberArray) >= static_conf.Basic.Union.MemberLimit {
		return rsp, code.PeopleIsFull
	}
	c := db.Coll(model.CoalitionApply{}.CollectionName())
	if c == nil {
		return rsp, code.DBError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cnt, _ := c.CountDocuments(ctx, bson.M{
		"server_id": db.ServerId(),
		"union_id":  req.Id,
		"rid":       role.RId,
		"state":     int8(protocol.UnionUntreated),
	})
	if cnt > 0 {
		return rsp, code.HasApply
	}
	apply := &model.CoalitionApply{
		ServerId: db.ServerId(),
		Id:       db.NextID(model.CoalitionApply{}.CollectionName()),
		RId:      role.RId,
		UnionId:  req.Id,
		Ctime:    time.Now(),
		State:    protocol.UnionUntreated,
	}
	if _, err := c.InsertOne(ctx, apply); err != nil {
		return rsp, code.DBError
	}
	apply.SyncExecute()
	return rsp, code.OK
}

func (p *UnionChild) verify(req *protocol.VerifyReq) (*protocol.VerifyRsp, int32) {
	rsp := &protocol.VerifyRsp{Id: req.Id, Decide: req.Decide}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	c := db.Coll(model.CoalitionApply{}.CollectionName())
	if c == nil {
		return rsp, code.DBError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	apply := &model.CoalitionApply{}
	err := c.FindOne(ctx, bson.M{
		"server_id": db.ServerId(),
		"id":        req.Id,
		"state":     int8(protocol.UnionUntreated),
	}).Decode(apply)
	if err != nil {
		return rsp, code.InvalidParam
	}
	targetRole, ok := mgr.RMgr.Get(apply.RId)
	if !ok {
		return rsp, code.RoleNotExist
	}
	u, ok := mgr.UnionMgr.Get(apply.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		return rsp, code.PermissionDenied
	}
	if len(u.MemberArray) >= static_conf.Basic.Union.MemberLimit {
		return rsp, code.PeopleIsFull
	}
	if mgr.RAttrMgr.IsHasUnion(apply.RId) {
		apply.State = req.Decide
		_, _ = c.UpdateOne(ctx,
			bson.M{"server_id": db.ServerId(), "id": apply.Id},
			bson.M{"$set": bson.M{"state": apply.State}})
		return rsp, code.UnionAlreadyHas
	}
	if req.Decide == protocol.UnionAdopt {
		u.MemberArray = append(u.MemberArray, apply.RId)
		if logic.Union != nil {
			logic.Union.MemberEnter(apply.RId, apply.UnionId)
		}
		u.SyncExecute()
		model.NewJoin(targetRole.NickName, apply.UnionId, role.RId, apply.RId)
	}
	apply.State = req.Decide
	_, _ = c.UpdateOne(ctx,
		bson.M{"server_id": db.ServerId(), "id": apply.Id},
		bson.M{"$set": bson.M{"state": apply.State}})
	return rsp, code.OK
}

func (p *UnionChild) member(req *protocol.MemberReq) (*protocol.MemberRsp, int32) {
	rsp := &protocol.MemberRsp{Id: req.Id, Members: make([]protocol.Member, 0)}
	u, ok := mgr.UnionMgr.Get(req.Id)
	if !ok {
		return rsp, code.UnionNotFound
	}
	for _, rid := range u.MemberArray {
		role, ok := mgr.RMgr.Get(rid)
		if !ok {
			continue
		}
		m := protocol.Member{RId: role.RId, Name: role.NickName}
		if main, ok := mgr.RCMgr.GetMainCity(role.RId); ok {
			m.X = main.X
			m.Y = main.Y
		}
		switch rid {
		case u.Chairman:
			m.Title = protocol.UnionChairman
		case u.ViceChairman:
			m.Title = protocol.UnionViceChairman
		default:
			m.Title = protocol.UnionCommon
		}
		rsp.Members = append(rsp.Members, m)
	}
	return rsp, code.OK
}

func (p *UnionChild) applyList(req *protocol.ApplyReq) (*protocol.ApplyRsp, int32) {
	rsp := &protocol.ApplyRsp{Id: req.Id, Applys: make([]protocol.ApplyItem, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	u, ok := mgr.UnionMgr.Get(req.Id)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		return rsp, code.OK
	}
	c := db.Coll(model.CoalitionApply{}.CollectionName())
	if c == nil {
		return rsp, code.DBError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := c.Find(ctx, bson.M{
		"server_id": db.ServerId(),
		"union_id":  req.Id,
		"state":     int8(protocol.UnionUntreated),
	})
	if err != nil {
		return rsp, code.DBError
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		a := &model.CoalitionApply{}
		if err := cur.Decode(a); err != nil {
			continue
		}
		r, ok := mgr.RMgr.Get(a.RId)
		if !ok {
			continue
		}
		rsp.Applys = append(rsp.Applys, protocol.ApplyItem{Id: a.Id, RId: a.RId, NickName: r.NickName})
	}
	return rsp, code.OK
}

func (p *UnionChild) exit(_ *protocol.ExitReq) (*protocol.ExitRsp, int32) {
	rsp := &protocol.ExitRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionNotFound
	}
	attr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(attr.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if u.Chairman == role.RId {
		return rsp, code.UnionNotAllowExit
	}
	for i, rid := range u.MemberArray {
		if rid == role.RId {
			u.MemberArray = append(u.MemberArray[:i], u.MemberArray[i+1:]...)
			break
		}
	}
	if u.ViceChairman == role.RId {
		u.ViceChairman = 0
	}
	if logic.Union != nil {
		logic.Union.MemberExit(role.RId)
	}
	u.SyncExecute()
	model.NewExit(role.NickName, u.Id, role.RId)
	return rsp, code.OK
}

func (p *UnionChild) dismiss(_ *protocol.DismissReq) (*protocol.DismissRsp, int32) {
	rsp := &protocol.DismissRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionNotFound
	}
	attr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(attr.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if u.Chairman != role.RId {
		return rsp, code.PermissionDenied
	}
	if logic.Union != nil {
		logic.Union.Dismiss(u.Id)
	}
	model.NewDismiss(role.NickName, u.Id, role.RId)
	return rsp, code.OK
}

func (p *UnionChild) notice(req *protocol.NoticeReq) (*protocol.NoticeRsp, int32) {
	rsp := &protocol.NoticeRsp{}
	u, ok := mgr.UnionMgr.Get(req.Id)
	if !ok {
		return rsp, code.UnionNotFound
	}
	rsp.Text = u.Notice
	return rsp, code.OK
}

func (p *UnionChild) modNotice(req *protocol.ModNoticeReq) (*protocol.ModNoticeRsp, int32) {
	rsp := &protocol.ModNoticeRsp{Text: req.Text}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if len(req.Text) > 200 {
		return rsp, code.ContentTooLong
	}
	if !mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionNotFound
	}
	attr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(attr.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		return rsp, code.PermissionDenied
	}
	rsp.Id = u.Id
	u.Notice = req.Text
	u.SyncExecute()
	model.NewModNotice(role.NickName, u.Id, role.RId)
	return rsp, code.OK
}

func (p *UnionChild) kick(req *protocol.KickReq) (*protocol.KickRsp, int32) {
	rsp := &protocol.KickRsp{RId: req.RId}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionNotFound
	}
	opAr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(opAr.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		return rsp, code.PermissionDenied
	}
	if role.RId == req.RId {
		return rsp, code.PermissionDenied
	}
	targetRole, ok := mgr.RMgr.Get(req.RId)
	if !ok {
		return rsp, code.RoleNotExist
	}
	target, ok := mgr.RAttrMgr.Get(req.RId)
	if !ok || target.UnionId != u.Id {
		return rsp, code.NotBelongUnion
	}
	for i, rid := range u.MemberArray {
		if rid == req.RId {
			u.MemberArray = append(u.MemberArray[:i], u.MemberArray[i+1:]...)
			break
		}
	}
	if u.ViceChairman == req.RId {
		u.ViceChairman = 0
	}
	if logic.Union != nil {
		logic.Union.MemberExit(req.RId)
	}
	target.UnionId = 0
	u.SyncExecute()
	model.NewKick(role.NickName, targetRole.NickName, u.Id, role.RId, target.RId)
	return rsp, code.OK
}

func (p *UnionChild) appoint(req *protocol.AppointReq) (*protocol.AppointRsp, int32) {
	rsp := &protocol.AppointRsp{RId: req.RId}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionNotFound
	}
	opAr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(opAr.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	if u.Chairman != role.RId {
		return rsp, code.PermissionDenied
	}
	targetRole, ok := mgr.RMgr.Get(req.RId)
	if !ok {
		return rsp, code.RoleNotExist
	}
	target, ok := mgr.RAttrMgr.Get(req.RId)
	if !ok || target.UnionId != u.Id {
		return rsp, code.NotBelongUnion
	}
	switch req.Title {
	case protocol.UnionViceChairman:
		u.ViceChairman = req.RId
		rsp.Title = req.Title
		u.SyncExecute()
		model.NewAppoint(role.NickName, targetRole.NickName, u.Id, role.RId, targetRole.RId, req.Title)
	case protocol.UnionCommon:
		if u.ViceChairman == req.RId {
			u.ViceChairman = 0
		}
		rsp.Title = req.Title
		model.NewAppoint(role.NickName, targetRole.NickName, u.Id, role.RId, targetRole.RId, req.Title)
	default:
		return rsp, code.InvalidParam
	}
	return rsp, code.OK
}

func (p *UnionChild) abdicate(req *protocol.AbdicateReq) (*protocol.AbdicateRsp, int32) {
	rsp := &protocol.AbdicateRsp{}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	if !mgr.RAttrMgr.IsHasUnion(role.RId) {
		return rsp, code.UnionNotFound
	}
	opAr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(opAr.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	targetRole, ok := mgr.RMgr.Get(req.RId)
	if !ok {
		return rsp, code.RoleNotExist
	}
	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		return rsp, code.PermissionDenied
	}
	target, ok := mgr.RAttrMgr.Get(req.RId)
	if !ok || target.UnionId != u.Id {
		return rsp, code.NotBelongUnion
	}
	if role.RId == u.Chairman {
		u.Chairman = req.RId
		if u.ViceChairman == req.RId {
			u.ViceChairman = 0
		}
		u.SyncExecute()
		model.NewAbdicate(role.NickName, targetRole.NickName, u.Id, role.RId, targetRole.RId, protocol.UnionChairman)
	} else if role.RId == u.ViceChairman {
		u.ViceChairman = req.RId
		u.SyncExecute()
		model.NewAbdicate(role.NickName, targetRole.NickName, u.Id, role.RId, targetRole.RId, protocol.UnionViceChairman)
	}
	return rsp, code.OK
}

func (p *UnionChild) info(req *protocol.InfoReq) (*protocol.InfoRsp, int32) {
	rsp := &protocol.InfoRsp{Id: req.Id}
	u, ok := mgr.UnionMgr.Get(req.Id)
	if !ok {
		return rsp, code.UnionNotFound
	}
	rsp.Info = u.ToProto().(protocol.Union)
	rsp.Info.Major = p.majorOf(u)
	return rsp, code.OK
}

func (p *UnionChild) unionLog(_ *protocol.LogReq) (*protocol.LogRsp, int32) {
	rsp := &protocol.LogRsp{Logs: make([]protocol.UnionLog, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	opAr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(opAr.UnionId)
	if !ok {
		return rsp, code.UnionNotFound
	}
	c := db.Coll(model.CoalitionLog{}.CollectionName())
	if c == nil {
		return rsp, code.DBError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := c.Find(ctx,
		bson.M{"server_id": db.ServerId(), "union_id": u.Id},
		options.Find().SetSort(bson.M{"ctime": -1}))
	if err != nil {
		return rsp, code.DBError
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		l := &model.CoalitionLog{}
		if err := cur.Decode(l); err != nil {
			continue
		}
		rsp.Logs = append(rsp.Logs, l.ToProto().(protocol.UnionLog))
	}
	return rsp, code.OK
}
