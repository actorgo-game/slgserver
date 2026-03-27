package component

import (
	"context"
	"fmt"
	"time"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const MongoComponentName = "mongo_component"

type MongoComponent struct {
	cfacade.Component
	dbMap map[string]*mongo.Database
}

func NewMongo() *MongoComponent {
	return &MongoComponent{
		dbMap: make(map[string]*mongo.Database),
	}
}

func (*MongoComponent) Name() string {
	return MongoComponentName
}

func (c *MongoComponent) Init() {
	mongoIDList := c.App().Settings().Get("mongo_id_list")
	if mongoIDList.LastError() != nil || mongoIDList.Size() < 1 {
		clog.Warn("[mongo] mongo_id_list not found in node settings")
		return
	}

	mongoConfig := cprofile.GetConfig("mongo")
	if mongoConfig.LastError() != nil {
		clog.Warn("[mongo] 'mongo' config not found in profile")
		return
	}

	for _, groupID := range mongoConfig.Keys() {
		dbGroup := mongoConfig.GetConfig(groupID)
		for i := 0; i < dbGroup.Size(); i++ {
			item := dbGroup.GetConfig(i)
			id := item.GetString("db_id")
			dbName := item.GetString("db_name")
			uri := item.GetString("uri")
			enable := item.GetBool("enable", true)
			timeout := time.Duration(item.GetInt64("timeout", 5)) * time.Second

			if !enable {
				continue
			}

			for _, key := range mongoIDList.Keys() {
				if mongoIDList.Get(key).ToString() != id {
					continue
				}

				db, err := createDatabase(uri, dbName, timeout)
				if err != nil {
					panic(fmt.Sprintf("[mongo] connect fail: db=%s, err=%v", dbName, err))
				}
				c.dbMap[id] = db
				clog.Info("[mongo] connected: id=%s, db=%s", id, dbName)
			}
		}
	}
}

func (c *MongoComponent) GetDb(id string) *mongo.Database {
	return c.dbMap[id]
}

func (c *MongoComponent) Collection(dbId, collName string) *mongo.Collection {
	db := c.dbMap[dbId]
	if db == nil {
		return nil
	}
	return db.Collection(collName)
}

func createDatabase(uri, dbName string, timeout time.Duration) (*mongo.Database, error) {
	opts := options.Client().ApplyURI(uri)
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client.Database(dbName), nil
}
