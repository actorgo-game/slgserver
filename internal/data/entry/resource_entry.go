package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ResourceEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache
	serverId int
}

func NewResourceEntry(coll *mongo.Collection, serverId int) *ResourceEntry {
	e := &ResourceEntry{
		coll:     coll,
		ridCache: newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "rid", Value: 1}}),
	)
	return e
}

func (e *ResourceEntry) FindByRId(rid int) (*model.RoleRes, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.(*model.RoleRes), nil
	}
	c, cancel := ctx()
	defer cancel()
	var res model.RoleRes
	err := e.coll.FindOne(c, bson.M{"server_id": e.serverId, "rid": rid}).Decode(&res)
	if err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, &res)
	return &res, nil
}

func (e *ResourceEntry) Upsert(res *model.RoleRes) error {
	res.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	filter := bson.M{"server_id": e.serverId, "rid": res.RId}
	_, err := e.coll.ReplaceOne(c, filter, res, replaceUpsert())
	if err == nil {
		e.ridCache.Put(res.RId, res)
	}
	return err
}
