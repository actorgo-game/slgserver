package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ArmyEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache // rid -> []*Army
	serverId int
}

func NewArmyEntry(coll *mongo.Collection, serverId int) *ArmyEntry {
	e := &ArmyEntry{
		coll:     coll,
		ridCache: newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		normalIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "rid", Value: 1}}),
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "id", Value: 1}}),
	)
	return e
}

func (e *ArmyEntry) FindByRId(rid int) ([]*model.Army, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.([]*model.Army), nil
	}
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "rid": rid})
	if err != nil {
		return nil, err
	}
	var armies []*model.Army
	if err = cursor.All(c, &armies); err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, armies)
	return armies, nil
}

func (e *ArmyEntry) Insert(a *model.Army) error {
	a.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, a)
	if err == nil {
		e.ridCache.Invalidate(a.RId)
	}
	return err
}

func (e *ArmyEntry) Update(a *model.Army) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "id": a.Id}, a)
	if err == nil {
		e.ridCache.Invalidate(a.RId)
	}
	return err
}

func (e *ArmyEntry) FindAll() ([]*model.Army, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId})
	if err != nil {
		return nil, err
	}
	var armies []*model.Army
	if err = cursor.All(c, &armies); err != nil {
		return nil, err
	}
	return armies, nil
}
