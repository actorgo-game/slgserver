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
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorWar struct {
	cactor.Base
}

func NewActorWar() *ActorWar       { return &ActorWar{} }
func (p *ActorWar) AliasID() string { return "war" }
func (p *ActorWar) OnInit()         { clog.Info("[ActorWar] initialized") }

func (p *ActorWar) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &WarChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type WarChild struct{ userActor }

func (p *WarChild) OnInit() {
	p.Remote().Register("report", p.report)
	p.Remote().Register("read", p.read)
}

func (p *WarChild) report(_ *protocol.WarReportReq) (*protocol.WarReportRsp, int32) {
	rsp := &protocol.WarReportRsp{List: make([]protocol.WarReport, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	c := db.Coll(model.WarReport{}.CollectionName())
	if c == nil {
		return rsp, code.DBError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{
		"server_id": db.ServerId(),
		"$or": []bson.M{
			{"a_rid": role.RId},
			{"d_rid": role.RId},
		},
	}
	opt := options.Find().SetSort(bson.M{"ctime": -1}).SetLimit(100)
	cur, err := c.Find(ctx, filter, opt)
	if err != nil {
		return rsp, code.DBError
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		w := &model.WarReport{}
		if err := cur.Decode(w); err != nil {
			continue
		}
		rsp.List = append(rsp.List, w.ToProto().(protocol.WarReport))
	}
	return rsp, code.OK
}

func (p *WarChild) read(req *protocol.WarReadReq) (*protocol.WarReadRsp, int32) {
	rsp := &protocol.WarReadRsp{Id: req.Id}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	c := db.Coll(model.WarReport{}.CollectionName())
	if c == nil {
		return rsp, code.DBError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if req.Id == 0 {
		_, _ = c.UpdateMany(ctx,
			bson.M{"server_id": db.ServerId(), "a_rid": role.RId},
			bson.M{"$set": bson.M{"a_is_read": true}})
		_, _ = c.UpdateMany(ctx,
			bson.M{"server_id": db.ServerId(), "d_rid": role.RId},
			bson.M{"$set": bson.M{"d_is_read": true}})
		return rsp, code.OK
	}
	wr := &model.WarReport{}
	err := c.FindOne(ctx, bson.M{"server_id": db.ServerId(), "id": int(req.Id)}).Decode(wr)
	if err != nil {
		return rsp, code.DBError
	}
	if wr.AttackRid == role.RId {
		_, _ = c.UpdateOne(ctx,
			bson.M{"server_id": db.ServerId(), "id": wr.Id},
			bson.M{"$set": bson.M{"a_is_read": true}})
		return rsp, code.OK
	}
	if wr.DefenseRid == role.RId {
		_, _ = c.UpdateOne(ctx,
			bson.M{"server_id": db.ServerId(), "id": wr.Id},
			bson.M{"$set": bson.M{"d_is_read": true}})
		return rsp, code.OK
	}
	return rsp, code.InvalidParam
}
