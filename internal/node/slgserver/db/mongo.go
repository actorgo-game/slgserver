package db

import (
	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/llr104/slgserver/internal/data/model"
)

// 包级 mongo 数据库句柄。run.Init() 启动时由 actorgo 的 cmongo 组件注入。
// model 包里所有异步落库 goroutine 都通过 Coll(name) 获取集合。
var (
	mdb      *mongo.Database
	dbId     string
	serverId int
)

// Setup 在 slgserver 节点启动时调用，把当前节点对应的 mongo 数据库句柄注入到本包。
//
// dbID 与 cluster.json 中 components.mongo.<group>.<id> 的 id 一致，例如 "slg_db"。
func Setup(id string, sid int) {
	dbId = id
	serverId = sid

	inst := cmongo.Instance()
	if inst == nil {
		clog.Panic("[slgserver/db] mongo component not registered")
		return
	}

	mdb = inst.GetDb(id)
	if mdb == nil {
		clog.Panic("[slgserver/db] mongo database not found: id=%s", id)
		return
	}
	model.SetServerId(sid)
	clog.Info("[slgserver/db] mongo ready: db=%s, serverId=%d", id, sid)
}

// DB 返回 mongo 数据库句柄。
func DB() *mongo.Database {
	return mdb
}

// Coll 返回指定集合。集合命名规范：保持与原 xorm 表名一致（去掉 _serverId 后缀）。
// serverId 维度通过文档内 server_id 字段隔离。
func Coll(name string) *mongo.Collection {
	if mdb == nil {
		return nil
	}
	return mdb.Collection(name)
}

// ServerId 返回当前节点对应的 serverId。
func ServerId() int {
	return serverId
}
