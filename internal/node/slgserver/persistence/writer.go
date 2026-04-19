package persistence

import (
	"context"
	"sync"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/db"
)

// 原 slgserver 的每个 model 都有自己的 channel + goroutine 异步落库。
// 这里聚合成一个全局 worker：
//
//   - SyncWriter 钩子把待持久化对象推到 inbox；
//   - 单 goroutine 串行从 inbox 取出，按对象具体类型分发到对应 collection 的 ReplaceOne(upsert)。
//
// 这样既满足"非阻塞 + 顺序一致"，又避免每个 model 都启动一个 goroutine。
const inboxSize = 4096

var (
	once   sync.Once
	inbox  chan any
	closed bool
)

// Setup 在 slgserver 节点启动时调用：
//
//	persistence.Setup()
//	model.SyncWriter = persistence.Submit
//
// 之后所有 model.SyncExecute() 都会走异步落库。
func Setup() {
	once.Do(func() {
		inbox = make(chan any, inboxSize)
		go run()
	})
}

// Submit 是 model.SyncWriter 钩子的实现。
func Submit(v any) {
	if closed {
		return
	}
	if inbox == nil {
		clog.Warn("[persistence] writer not ready, drop one record")
		return
	}
	select {
	case inbox <- v:
	default:
		clog.Warn("[persistence] inbox full, drop one record")
	}
}

func run() {
	for v := range inbox {
		dispatch(v)
	}
}

// dispatch 根据具体类型决定写哪张表 + 用什么 filter。
// 与原 slgserver 各 model 的 SyncExecute 一一对应。
func dispatch(v any) {
	defer func() {
		if r := recover(); r != nil {
			clog.Warn("[persistence] panic recovered: %v", r)
		}
	}()

	switch m := v.(type) {
	case *model.Role:
		upsertByID(model.Role{}.CollectionName(), bson.M{"rid": m.RId}, m)
	case *model.RoleAttribute:
		upsertByID(model.RoleAttribute{}.CollectionName(), bson.M{"rid": m.RId}, m)
	case *model.RoleRes:
		upsertByID(model.RoleRes{}.CollectionName(), bson.M{"rid": m.RId}, m)
	case *model.MapRoleCity:
		upsertByID(model.MapRoleCity{}.CollectionName(), bson.M{"cityId": m.CityId}, m)
	case *model.MapRoleBuild:
		upsertByID(model.MapRoleBuild{}.CollectionName(), bson.M{"id": m.Id}, m)
	case *model.CityFacility:
		upsertByID(model.CityFacility{}.CollectionName(), bson.M{"cityId": m.CityId}, m)
	case *model.General:
		upsertByID(model.General{}.CollectionName(), bson.M{"id": m.Id}, m)
	case *model.Skill:
		upsertByID(model.Skill{}.CollectionName(), bson.M{"id": m.Id}, m)
	case *model.Army:
		upsertByID(model.Army{}.CollectionName(), bson.M{"id": m.Id}, m)
	case *model.Coalition:
		upsertByID(model.Coalition{}.CollectionName(), bson.M{"id": m.Id}, m)
	case *model.CoalitionApply:
		upsertByID(model.CoalitionApply{}.CollectionName(), bson.M{"id": m.Id}, m)
	case *model.CoalitionLog:
		insertOne(model.CoalitionLog{}.CollectionName(), m)
	case *model.WarReport:
		insertOne(model.WarReport{}.CollectionName(), m)
	default:
		clog.Warn("[persistence] unknown type: %T", v)
	}
}

func upsertByID(coll string, key bson.M, doc any) {
	c := db.Coll(coll)
	if c == nil {
		clog.Warn("[persistence] collection nil: %s", coll)
		return
	}
	key["server_id"] = db.ServerId()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := c.ReplaceOne(ctx, key, doc, options.Replace().SetUpsert(true))
	if err != nil {
		clog.Warn("[persistence] ReplaceOne error: coll=%s, key=%v, err=%v", coll, key, err)
	}
}

func insertOne(coll string, doc any) {
	c := db.Coll(coll)
	if c == nil {
		clog.Warn("[persistence] collection nil: %s", coll)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := c.InsertOne(ctx, doc)
	if err != nil {
		clog.Warn("[persistence] InsertOne error: coll=%s, err=%v", coll, err)
	}
}

// 用于在 controller / mgr 里直接执行同步写（少数场景，例如 Insert 完成后立即拿到 id 再 Push 给 client）。
var _ = mongo.ErrNoDocuments
