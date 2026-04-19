package db

import (
	"context"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// counters 集合按 (server_id, name) 维度自增，等价于原 xorm 的 autoincr 主键。
//
// 文档结构：
//
//	{ server_id: 1, name: "armies", seq: 1234 }
//
// 调用方按 model 的 CollectionName() 作为 name，例如：
//
//	id := db.NextID(model.Army{}.CollectionName())
const counterColl = "counters"

// NextID 返回下一个自增 id。失败返回 0 并打 warn 日志（与原 xorm Insert 失败的语义一致）。
func NextID(name string) int {
	if mdb == nil {
		clog.Warn("[idgen] mongo not ready: name=%s", name)
		return 0
	}
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"server_id": serverId, "name": name}
	update := bson.M{"$inc": bson.M{"seq": int64(1)}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var doc struct {
		Seq int64 `bson:"seq"`
	}
	err := mdb.Collection(counterColl).FindOneAndUpdate(c, filter, update, opts).Decode(&doc)
	if err != nil {
		clog.Warn("[idgen] FindOneAndUpdate error: name=%s, err=%v", name, err)
		return 0
	}
	return int(doc.Seq)
}
