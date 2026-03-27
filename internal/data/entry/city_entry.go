package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CityEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache
	serverId int
}

func NewCityEntry(coll *mongo.Collection, serverId int) *CityEntry {
	e := &CityEntry{
		coll:     coll,
		ridCache: newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "cityId", Value: 1}}),
		normalIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "rid", Value: 1}}),
	)
	return e
}

func (e *CityEntry) FindByRId(rid int) ([]*model.MapRoleCity, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.([]*model.MapRoleCity), nil
	}
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "rid": rid})
	if err != nil {
		return nil, err
	}
	var cities []*model.MapRoleCity
	if err = cursor.All(c, &cities); err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, cities)
	return cities, nil
}

func (e *CityEntry) Insert(city *model.MapRoleCity) error {
	city.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, city)
	if err == nil {
		e.ridCache.Invalidate(city.RId)
	}
	return err
}

func (e *CityEntry) Update(city *model.MapRoleCity) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "cityId": city.CityId}, city)
	if err == nil {
		e.ridCache.Invalidate(city.RId)
	}
	return err
}

func (e *CityEntry) FindAll() ([]*model.MapRoleCity, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId})
	if err != nil {
		return nil, err
	}
	var cities []*model.MapRoleCity
	if err = cursor.All(c, &cities); err != nil {
		return nil, err
	}
	return cities, nil
}
