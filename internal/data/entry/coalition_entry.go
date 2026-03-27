package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CoalitionEntry struct {
	coll     *mongo.Collection
	idCache  cache.Cache
	serverId int
}

func NewCoalitionEntry(coll *mongo.Collection, serverId int) *CoalitionEntry {
	e := &CoalitionEntry{
		coll:     coll,
		idCache:  newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "id", Value: 1}}),
	)
	return e
}

func (e *CoalitionEntry) FindById(id int) (*model.Coalition, error) {
	if v, ok := e.idCache.GetIfPresent(id); ok {
		return v.(*model.Coalition), nil
	}
	c, cancel := ctx()
	defer cancel()
	var co model.Coalition
	err := e.coll.FindOne(c, bson.M{"server_id": e.serverId, "id": id}).Decode(&co)
	if err != nil {
		return nil, err
	}
	e.idCache.Put(id, &co)
	return &co, nil
}

func (e *CoalitionEntry) FindAllRunning() ([]*model.Coalition, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "state": model.UnionRunning})
	if err != nil {
		return nil, err
	}
	var coalitions []*model.Coalition
	if err = cursor.All(c, &coalitions); err != nil {
		return nil, err
	}
	for _, co := range coalitions {
		e.idCache.Put(co.Id, co)
	}
	return coalitions, nil
}

func (e *CoalitionEntry) Insert(co *model.Coalition) error {
	co.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, co)
	if err == nil {
		e.idCache.Put(co.Id, co)
	}
	return err
}

func (e *CoalitionEntry) Update(co *model.Coalition) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "id": co.Id}, co)
	if err == nil {
		e.idCache.Put(co.Id, co)
	}
	return err
}

type CoalitionApplyEntry struct {
	coll     *mongo.Collection
	serverId int
}

func NewCoalitionApplyEntry(coll *mongo.Collection, serverId int) *CoalitionApplyEntry {
	return &CoalitionApplyEntry{coll: coll, serverId: serverId}
}

func (e *CoalitionApplyEntry) FindByUnionId(unionId int) ([]*model.CoalitionApply, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "union_id": unionId, "state": 0})
	if err != nil {
		return nil, err
	}
	var applies []*model.CoalitionApply
	err = cursor.All(c, &applies)
	return applies, err
}

func (e *CoalitionApplyEntry) Insert(a *model.CoalitionApply) error {
	a.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, a)
	return err
}

func (e *CoalitionApplyEntry) Update(a *model.CoalitionApply) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "id": a.Id}, a)
	return err
}

type CoalitionLogEntry struct {
	coll     *mongo.Collection
	serverId int
}

func NewCoalitionLogEntry(coll *mongo.Collection, serverId int) *CoalitionLogEntry {
	return &CoalitionLogEntry{coll: coll, serverId: serverId}
}

func (e *CoalitionLogEntry) FindByUnionId(unionId int) ([]*model.CoalitionLog, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "union_id": unionId})
	if err != nil {
		return nil, err
	}
	var logs []*model.CoalitionLog
	err = cursor.All(c, &logs)
	return logs, err
}

func (e *CoalitionLogEntry) Insert(l *model.CoalitionLog) error {
	l.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, l)
	return err
}
