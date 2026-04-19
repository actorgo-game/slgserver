package army

import (
	"github.com/llr104/slgserver/internal/node/slgserver/global"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/util"
)

var ViewWidth = 5
var ViewHeight = 5

// ArmyIsInView 是否在视野范围内（与原 slgserver 同语义）。
func ArmyIsInView(rid, x, y int) bool {
	unionId := mgr.UnionId(rid)
	for i := util.MaxInt(x-ViewWidth, 0); i < util.MinInt(x+ViewWidth, global.MapWith); i++ {
		for j := util.MaxInt(y-ViewHeight, 0); j < util.MinInt(y+ViewHeight, global.MapHeight); j++ {
			if build, ok := mgr.RBMgr.PositionBuild(i, j); ok {
				tUnionId := mgr.UnionId(build.RId)
				if (tUnionId != 0 && unionId == tUnionId) || build.RId == rid {
					return true
				}
			}
			if city, ok := mgr.RCMgr.PositionCity(i, j); ok {
				tUnionId := mgr.UnionId(city.RId)
				if (tUnionId != 0 && unionId == tUnionId) || city.RId == rid {
					return true
				}
			}
		}
	}
	return false
}
