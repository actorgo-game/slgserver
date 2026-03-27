package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type GeneralEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache // rid -> []*General
	serverId int
}

func NewGeneralEntry(coll *mongo.Collection, serverId int) *GeneralEntry {
	e := &GeneralEntry{
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

func (e *GeneralEntry) FindByRId(rid int) ([]*model.General, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.([]*model.General), nil
	}
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "rid": rid, "state": model.GeneralNormal})
	if err != nil {
		return nil, err
	}
	var generals []*model.General
	if err = cursor.All(c, &generals); err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, generals)
	return generals, nil
}

func (e *GeneralEntry) Insert(g *model.General) error {
	g.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, g)
	if err == nil {
		e.ridCache.Invalidate(g.RId)
	}
	return err
}

func (e *GeneralEntry) Update(g *model.General) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "id": g.Id}, g)
	if err == nil {
		e.ridCache.Invalidate(g.RId)
	}
	return err
}

func (e *GeneralEntry) InsertMany(generals []*model.General) error {
	if len(generals) == 0 {
		return nil
	}
	docs := make([]interface{}, len(generals))
	for i, g := range generals {
		g.ServerId = e.serverId
		docs[i] = g
	}
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertMany(c, docs)
	return err
}
