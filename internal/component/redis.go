package component

import (
	"context"
	"fmt"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
	"github.com/go-redis/redis/v8"
)

const RedisComponentName = "redis_component"

type RedisComponent struct {
	cfacade.Component
	client *redis.Client
}

func NewRedis() *RedisComponent {
	return &RedisComponent{}
}

func (*RedisComponent) Name() string {
	return RedisComponentName
}

func (c *RedisComponent) Init() {
	redisCfg := cprofile.GetConfig("redis")
	if redisCfg.LastError() != nil {
		clog.Warn("[redis] config not found in profile")
		return
	}

	addr := redisCfg.GetString("address", "127.0.0.1:6379")
	password := redisCfg.GetString("password", "")
	db := redisCfg.GetInt("db", 0)

	c.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := c.client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("[redis] connect fail: addr=%s, err=%v", addr, err))
	}

	clog.Info("[redis] connected: addr=%s, db=%d", addr, db)
}

func (c *RedisComponent) OnStop() {
	if c.client != nil {
		_ = c.client.Close()
	}
}

func (c *RedisComponent) Client() *redis.Client {
	return c.client
}
