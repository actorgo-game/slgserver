package entry

import (
	"github.com/goburrow/cache"
	"github.com/llr104/slgserver/internal/data/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AccountEntry struct {
	coll      *mongo.Collection
	nameCache cache.Cache
}

func NewAccountEntry(coll *mongo.Collection) *AccountEntry {
	e := &AccountEntry{
		coll:      coll,
		nameCache: newCache(),
	}
	ensureIndexes(coll, uniqueIndex(bson.D{{Key: "username", Value: 1}}))
	return e
}

func (e *AccountEntry) FindByUsername(username string) (*model.Account, error) {
	if v, ok := e.nameCache.GetIfPresent(username); ok {
		return v.(*model.Account), nil
	}
	c, cancel := ctx()
	defer cancel()
	var acc model.Account
	err := e.coll.FindOne(c, bson.M{"username": username}).Decode(&acc)
	if err != nil {
		return nil, err
	}
	e.nameCache.Put(username, &acc)
	return &acc, nil
}

func (e *AccountEntry) Insert(acc *model.Account) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.InsertOne(c, acc)
	if err == nil {
		e.nameCache.Put(acc.Username, acc)
	}
	return err
}

func (e *AccountEntry) Update(acc *model.Account) error {
	c, cancel := ctx()
	defer cancel()
	_, err := e.coll.ReplaceOne(c, bson.M{"username": acc.Username}, acc)
	if err == nil {
		e.nameCache.Put(acc.Username, acc)
	}
	return err
}
