package entry

import (
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type WarReportEntry struct {
	coll     *mongo.Collection
	serverId int
}

func NewWarReportEntry(coll *mongo.Collection, serverId int) *WarReportEntry {
	e := &WarReportEntry{coll: coll, serverId: serverId}
	ensureIndexes(coll,
		normalIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "a_rid", Value: 1}}),
		normalIndex(bson.D{{Key: "server_id", Value: 1}, {Key: "d_rid", Value: 1}}),
		normalIndex(bson.D{{Key: "ctime", Value: -1}}),
	)
	return e
}

func (e *WarReportEntry) FindByRId(rid int, limit int64) ([]*model.WarReport, error) {
	c, cancel := ctx()
	defer cancel()
	filter := bson.M{
		"server_id": e.serverId,
		"$or": []bson.M{
			{"a_rid": rid},
			{"d_rid": rid},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "ctime", Value: -1}}).SetLimit(limit)
	cursor, err := e.coll.Find(c, filter, opts)
	if err != nil {
		return nil, err
	}
	var reports []*model.WarReport
	err = cursor.All(c, &reports)
	return reports, err
}

func (e *WarReportEntry) Insert(r *model.WarReport) error {
	r.ServerId = e.serverId
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, r)
	return err
}

func (e *WarReportEntry) MarkRead(id int, isAttack bool) error {
	c, cancel := ctx()
	defer cancel()
	field := "a_is_read"
	if !isAttack {
		field = "d_is_read"
	}
	_, err := e.coll.UpdateOne(c, bson.M{"server_id": e.serverId, "id": id}, bson.M{"$set": bson.M{field: true}})
	return err
}
