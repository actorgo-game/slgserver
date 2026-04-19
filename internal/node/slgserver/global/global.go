package global

import (
	cprofile "github.com/actorgo-game/actorgo/profile"
)

// MapWith / MapHeight 由 mgr.NMMgr.Load() 根据 mapRes_*.json 实际尺寸覆盖。
var MapWith = 200
var MapHeight = 200

func ToPosition(x, y int) int {
	return x + MapHeight*y
}

// IsDev 读取 actorgo profile 的 is_dev 配置。
func IsDev() bool {
	return cprofile.GetConfig("is_dev").ToBool()
}
