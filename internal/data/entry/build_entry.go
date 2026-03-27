package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BuildEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache
	serverId int
}

func NewBuildEntry(coll *mongo.Collection, serverId int) *BuildEntry {
	e := &BuildEntry{
		coll:     coll,
		ridCache: newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "id", Value: 1}}),
		normalIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "rid", Value: 1}}),
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "x", Value: 1}, {Key: "y", Value: 1}}),
	)
	return e
}

func (e *BuildEntry) FindByRId(rid int) ([]*model.MapRoleBuild, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.([]*model.MapRoleBuild), nil
	}
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "rid": rid})
	if err != nil {
		return nil, err
	}
	var builds []*model.MapRoleBuild
	if err = cursor.All(c, &builds); err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, builds)
	return builds, nil
}

func (e *BuildEntry) FindAll() ([]*model.MapRoleBuild, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId})
	if err != nil {
		return nil, err
	}
	var builds []*model.MapRoleBuild
	if err = cursor.All(c, &builds); err != nil {
		return nil, err
	}
	return builds, nil
}

func (e *BuildEntry) Insert(b *model.MapRoleBuild) error {
	b.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, b)
	if err == nil {
		e.ridCache.Invalidate(b.RId)
	}
	return err
}

func (e *BuildEntry) Update(b *model.MapRoleBuild) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "id": b.Id}, b)
	if err == nil {
		e.ridCache.Invalidate(b.RId)
	}
	return err
}

func (e *BuildEntry) Delete(id int) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.DeleteOne(c, bson.M{"server_id": e.serverId, "id": id})
	return err
}
