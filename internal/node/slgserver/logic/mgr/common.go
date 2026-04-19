package mgr

import (
	"context"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/llr104/slgserver/internal/node/slgserver/db"
)

// 与原 slgserver/logic/mgr/national_map_mgr.go 中常量一致：扫描的半径。
const (
	ScanWith   = 3
	ScanHeight = 3
)

// Distance 计算两点欧氏距离。
func Distance(begX, begY, endX, endY int) float64 {
	w := math.Abs(float64(endX - begX))
	h := math.Abs(float64(endY - begY))
	return math.Sqrt(w*w + h*h)
}

// TravelTime 与原 slgserver 同语义：返回毫秒。
func TravelTime(speed, begX, begY, endX, endY int) int {
	dis := Distance(begX, begY, endX, endY)
	t := dis / float64(speed) * 100000000
	return int(t)
}

// ctx 提供一个统一的 5 秒超时上下文。
func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// withServer 在所有读写 filter 上自动加上 server_id 维度。
func withServer(filter bson.M) bson.M {
	if filter == nil {
		filter = bson.M{}
	}
	filter["server_id"] = db.ServerId()
	return filter
}

// coll 读取指定 collection；nil 时返回 nil（调用方需判空）。
func coll(name string) *mongo.Collection {
	return db.Coll(name)
}
