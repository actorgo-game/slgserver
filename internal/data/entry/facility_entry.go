package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FacilityEntry struct {
	coll      *mongo.Collection
	cityCache cache.Cache
	serverId  int
}

func NewFacilityEntry(coll *mongo.Collection, serverId int) *FacilityEntry {
	e := &FacilityEntry{
		coll:      coll,
		cityCache: newCache(),
		serverId:  serverId,
	}
	ensureIndexes(coll,
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "cityId", Value: 1}}),
	)
	return e
}

func (e *FacilityEntry) FindByCityId(cityId int) (*model.CityFacility, error) {
	if v, ok := e.cityCache.GetIfPresent(cityId); ok {
		return v.(*model.CityFacility), nil
	}
	c, cancel := ctx()
	defer cancel()
	var f model.CityFacility
	err := e.coll.FindOne(c, bson.M{"server_id": e.serverId, "cityId": cityId}).Decode(&f)
	if err != nil {
		return nil, err
	}
	e.cityCache.Put(cityId, &f)
	return &f, nil
}

func (e *FacilityEntry) FindByRId(rid int) ([]*model.CityFacility, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "rid": rid})
	if err != nil {
		return nil, err
	}
	var list []*model.CityFacility
	if err = cursor.All(c, &list); err != nil {
		return nil, err
	}
	for _, f := range list {
		e.cityCache.Put(f.CityId, f)
	}
	return list, nil
}

func (e *FacilityEntry) Upsert(f *model.CityFacility) error {
	f.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	filter := bson.M{"server_id": e.serverId, "cityId": f.CityId}
	_, err := e.coll.ReplaceOne(c, filter, f, replaceUpsert())
	if err == nil {
		e.cityCache.Put(f.CityId, f)
	}
	return err
}

func replaceUpsert() *options.ReplaceOptionsBuilder {
	return options.Replace().SetUpsert(true)
}
