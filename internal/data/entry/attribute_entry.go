package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AttributeEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache
	serverId int
}

func NewAttributeEntry(coll *mongo.Collection, serverId int) *AttributeEntry {
	e := &AttributeEntry{
		coll:     coll,
		ridCache: newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "rid", Value: 1}}),
	)
	return e
}

func (e *AttributeEntry) FindByRId(rid int) (*model.RoleAttribute, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.(*model.RoleAttribute), nil
	}
	c, cancel := ctx()
	defer cancel()
	var attr model.RoleAttribute
	err := e.coll.FindOne(c, bson.M{"server_id": e.serverId, "rid": rid}).Decode(&attr)
	if err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, &attr)
	return &attr, nil
}

func (e *AttributeEntry) Upsert(attr *model.RoleAttribute) error {
	attr.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	filter := bson.M{"server_id": e.serverId, "rid": attr.RId}
	_, err := e.coll.ReplaceOne(c, filter, attr, replaceUpsert())
	if err == nil {
		e.ridCache.Put(attr.RId, attr)
	}
	return err
}
