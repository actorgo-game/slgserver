package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RoleEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache
	uidCache cache.Cache
	serverId int
}

func NewRoleEntry(coll *mongo.Collection, serverId int) *RoleEntry {
	e := &RoleEntry{
		coll:     coll,
		ridCache: newCache(),
		uidCache: newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		uniqueIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "rid", Value: 1}}),
		normalIndex(bson.D{{Key: "uid", Value: 1}}),
	)
	return e
}

func (e *RoleEntry) FindByRId(rid int) (*model.Role, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.(*model.Role), nil
	}
	c, cancel := ctx()
	defer cancel()
	var role model.Role
	err := e.coll.FindOne(c, bson.M{"server_id": e.serverId, "rid": rid}).Decode(&role)
	if err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, &role)
	return &role, nil
}

func (e *RoleEntry) FindByUId(uid int) (*model.Role, error) {
	if v, ok := e.uidCache.GetIfPresent(uid); ok {
		return v.(*model.Role), nil
	}
	c, cancel := ctx()
	defer cancel()
	var role model.Role
	err := e.coll.FindOne(c, bson.M{"server_id": e.serverId, "uid": uid}).Decode(&role)
	if err != nil {
		return nil, err
	}
	e.ridCache.Put(role.RId, &role)
	e.uidCache.Put(uid, &role)
	return &role, nil
}

func (e *RoleEntry) Insert(role *model.Role) error {
	role.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, role)
	if err == nil {
		e.ridCache.Put(role.RId, role)
		e.uidCache.Put(role.UId, role)
	}
	return err
}

func (e *RoleEntry) Update(role *model.Role) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "rid": role.RId}, role)
	if err == nil {
		e.ridCache.Put(role.RId, role)
	}
	return err
}

func (e *RoleEntry) FindAll() ([]*model.Role, error) {
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId})
	if err != nil {
		return nil, err
	}
	var roles []*model.Role
	if err = cursor.All(c, &roles); err != nil {
		return nil, err
	}
	for _, r := range roles {
		e.ridCache.Put(r.RId, r)
		e.uidCache.Put(r.UId, r)
	}
	return roles, nil
}
