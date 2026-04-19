package union

import (
	"sync"

	clog "github.com/actorgo-game/actorgo/logger"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
)

// 与原 slgserver 同名的 helper（实际实现已在 mgr 包，这里只做转发，
// 为兼容老代码与 controller 调用风格保留）。

func GetUnionId(rid int) int           { return mgr.UnionId(rid) }
func GetUnionName(unionId int) string  { return mgr.UnionName(unionId) }
func GetParentId(rid int) int          { return mgr.ParentId(rid) }
func GetMainMembers(unionId int) []int { return mgr.MainMembers(unionId) }

var _unionLogic *UnionLogic

func Instance() *UnionLogic {
	if _unionLogic == nil {
		_unionLogic = newUnionLogic()
	}
	return _unionLogic
}

type UnionLogic struct {
	mutex    sync.RWMutex
	children map[int]map[int]int // key:unionId, key&value: child rid
}

func newUnionLogic() *UnionLogic {
	c := &UnionLogic{children: make(map[int]map[int]int)}
	c.init()
	return c
}

func (this *UnionLogic) init() {
	for _, attr := range mgr.RAttrMgr.List() {
		if attr.ParentId != 0 {
			this.PutChild(attr.ParentId, attr.RId)
		}
	}
}

func (this *UnionLogic) MemberEnter(rid, unionId int) {
	attr, ok := mgr.RAttrMgr.TryCreate(rid)
	if ok {
		attr.UnionId = unionId
		if attr.ParentId == unionId {
			this.DelChild(unionId, attr.RId)
		}
	} else {
		clog.Warn("[UnionLogic] EnterUnion not found roleAttribute rid=%d", rid)
	}
	if rcs, ok := mgr.RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.SyncExecute()
		}
	}
}

func (this *UnionLogic) MemberExit(rid int) {
	if ra, ok := mgr.RAttrMgr.Get(rid); ok {
		ra.UnionId = 0
	}
	if rcs, ok := mgr.RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.SyncExecute()
		}
	}
}

// Dismiss 解散联盟。
func (this *UnionLogic) Dismiss(unionId int) {
	u, ok := mgr.UnionMgr.Get(unionId)
	if !ok {
		return
	}
	mgr.UnionMgr.Remove(unionId)
	for _, rid := range u.MemberArray {
		this.MemberExit(rid)
		this.DelUnionAllChild(unionId)
	}
	u.State = model.UnionDismiss
	u.MemberArray = []int{}
	u.SyncExecute()
}

func (this *UnionLogic) PutChild(unionId, rid int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _, ok := this.children[unionId]; !ok {
		this.children[unionId] = make(map[int]int)
	}
	this.children[unionId][rid] = rid
}

func (this *UnionLogic) DelChild(unionId, rid int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if children, ok := this.children[unionId]; ok {
		if attr, ok := mgr.RAttrMgr.Get(rid); ok {
			attr.ParentId = 0
			attr.SyncExecute()
		}
		delete(children, rid)
	}
}

func (this *UnionLogic) DelUnionAllChild(unionId int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if children, ok := this.children[unionId]; ok {
		for _, child := range children {
			if attr, ok := mgr.RAttrMgr.Get(child); ok {
				attr.ParentId = 0
				attr.SyncExecute()
			}
			if city, ok := mgr.RCMgr.GetMainCity(child); ok {
				city.SyncExecute()
			}
		}
		delete(this.children, unionId)
	}
}
