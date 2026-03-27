package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SkillEntry struct {
	coll     *mongo.Collection
	ridCache cache.Cache
	serverId int
}

func NewSkillEntry(coll *mongo.Collection, serverId int) *SkillEntry {
	e := &SkillEntry{
		coll:     coll,
		ridCache: newCache(),
		serverId: serverId,
	}
	ensureIndexes(coll,
		normalIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "rid", Value: 1}}),
	)
	return e
}

func (e *SkillEntry) FindByRId(rid int) ([]*model.Skill, error) {
	if v, ok := e.ridCache.GetIfPresent(rid); ok {
		return v.([]*model.Skill), nil
	}
	c, cancel := ctx()
	defer cancel()
	cursor, err := e.coll.Find(c, bson.M{"server_id": e.serverId, "rid": rid})
	if err != nil {
		return nil, err
	}
	var skills []*model.Skill
	if err = cursor.All(c, &skills); err != nil {
		return nil, err
	}
	e.ridCache.Put(rid, skills)
	return skills, nil
}

func (e *SkillEntry) Insert(s *model.Skill) error {
	s.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, s)
	if err == nil {
		e.ridCache.Invalidate(s.RId)
	}
	return err
}

func (e *SkillEntry) Update(s *model.Skill) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"server_id": e.serverId, "id": s.Id}, s)
	if err == nil {
		e.ridCache.Invalidate(s.RId)
	}
	return err
}
