package entry

import (
	"context"
	"time"

	"github.com/goburrow/cache"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const defaultTimeout = 5 * time.Second

func newCache() cache.Cache {
	return cache.New(
		cache.WithMaximumSize(0),
		cache.WithExpireAfterAccess(60*time.Minute),
	)
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

func ensureIndexes(coll *mongo.Collection, indexes ...mongo.IndexModel) {
	c, cancel := ctx()
	defer cancel()
	_, _ = coll.Indexes().CreateMany(c, indexes)
}

func uniqueIndex(keys bson.D) mongo.IndexModel {
	return mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(true),
	}
}

func normalIndex(keys bson.D) mongo.IndexModel {
	return mongo.IndexModel{Keys: keys}
}
