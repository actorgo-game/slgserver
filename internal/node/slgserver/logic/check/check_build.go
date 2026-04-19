package check

import (
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/util"
)

// IsCanArrive 与原 slgserver 同语义。
func IsCanArrive(x, y, rid int) bool {
	radius := 0
	unionId := mgr.UnionId(rid)
	if b, ok := mgr.RBMgr.PositionBuild(x, y); ok {
		radius = b.CellRadius()
	}
	if c, ok := mgr.RCMgr.PositionCity(x, y); ok {
		radius = c.CellRadius()
	}

	for tx := x - 10; tx <= x+10; tx++ {
		for ty := y - 10; ty <= y+10; ty++ {
			if b1, ok := mgr.RBMgr.PositionBuild(tx, ty); ok {
				absX := util.AbsInt(x - tx)
				absY := util.AbsInt(y - ty)
				if absX <= radius+b1.CellRadius()+1 && absY <= radius+b1.CellRadius()+1 {
					unionId1 := mgr.UnionId(b1.RId)
					parentId := mgr.ParentId(b1.RId)
					if b1.RId == rid || (unionId != 0 && (unionId == unionId1 || unionId == parentId)) {
						return true
					}
				}
			}

			if c1, ok := mgr.RCMgr.PositionCity(tx, ty); ok {
				absX := util.AbsInt(x - tx)
				absY := util.AbsInt(y - ty)
				if absX <= radius+c1.CellRadius()+1 && absY <= radius+c1.CellRadius()+1 {
					unionId1 := mgr.UnionId(c1.RId)
					parentId := mgr.ParentId(c1.RId)
					if c1.RId == rid || (unionId != 0 && (unionId == unionId1 || unionId == parentId)) {
						return true
					}
				}
			}
		}
	}
	return false
}

func IsCanDefend(x, y, rid int) bool {
	unionId := mgr.UnionId(rid)
	if b, ok := mgr.RBMgr.PositionBuild(x, y); ok {
		tUnionId := mgr.UnionId(b.RId)
		tParentId := mgr.ParentId(b.RId)
		if b.RId == rid {
			return true
		} else if tUnionId > 0 {
			return tUnionId == unionId
		} else if tParentId > 0 {
			return tParentId == unionId
		}
	}
	if c, ok := mgr.RCMgr.PositionCity(x, y); ok {
		tUnionId := mgr.UnionId(c.RId)
		tParentId := mgr.ParentId(c.RId)
		if c.RId == rid {
			return true
		} else if tUnionId > 0 {
			return tUnionId == unionId
		} else if tParentId > 0 {
			return tParentId == unionId
		}
	}
	return false
}

func IsWarFree(x, y int) bool {
	if b, ok := mgr.RBMgr.PositionBuild(x, y); ok {
		return b.IsWarFree()
	}
	if c, ok := mgr.RCMgr.PositionCity(x, y); ok && mgr.ParentId(c.RId) > 0 {
		return c.IsWarFree()
	}
	return false
}
