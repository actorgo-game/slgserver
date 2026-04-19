package mgr

import (
	"sync"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/data/model"
)

type roleAttributeMgr struct {
	mutex     sync.RWMutex
	attribute map[int]*model.RoleAttribute // key: rid
}

var RAttrMgr = &roleAttributeMgr{
	attribute: make(map[int]*model.RoleAttribute),
}

// Load 从 mongo 把所有 RoleAttribute 拉到内存，并按联盟列表回填 UnionId。
// 与原 slgserver 的 Load() 同语义；调用时机：必须在 UnionMgr.Load() 之后。
func (this *roleAttributeMgr) Load() {
	c := coll(model.RoleAttribute{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()

	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[RAttrMgr] Load find err=%v", err)
		return
	}
	var rows []*model.RoleAttribute
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[RAttrMgr] Load decode err=%v", err)
		return
	}
	for _, v := range rows {
		this.attribute[v.RId] = v
	}

	// 根据联盟列表回填 UnionId
	for _, u := range UnionMgr.List() {
		for _, rid := range u.MemberArray {
			if attr, ok := this.attribute[rid]; ok {
				attr.UnionId = u.Id
			} else {
				if a := this.create(rid); a != nil {
					a.UnionId = u.Id
				}
			}
		}
	}
}

func (this *roleAttributeMgr) Get(rid int) (*model.RoleAttribute, bool) {
	this.mutex.RLock()
	r, ok := this.attribute[rid]
	this.mutex.RUnlock()
	return r, ok
}

func (this *roleAttributeMgr) TryCreate(rid int) (*model.RoleAttribute, bool) {
	if attr, ok := this.Get(rid); ok {
		return attr, true
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	a := this.create(rid)
	return a, a != nil
}

func (this *roleAttributeMgr) create(rid int) *model.RoleAttribute {
	c := coll(model.RoleAttribute{}.CollectionName())
	if c == nil {
		return nil
	}
	attr := &model.RoleAttribute{
		Id:       nextID(model.RoleAttribute{}.CollectionName()),
		ServerId: serverID(),
		RId:      rid,
	}
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, attr); err != nil {
		clog.Error("[RAttrMgr] insert RoleAttribute err=%v rid=%d", err, rid)
		return nil
	}
	this.attribute[rid] = attr
	return attr
}

func (this *roleAttributeMgr) IsHasUnion(rid int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if r, ok := this.attribute[rid]; ok {
		return r.UnionId != 0
	}
	return false
}

func (this *roleAttributeMgr) UnionId(rid int) int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if r, ok := this.attribute[rid]; ok {
		return r.UnionId
	}
	return 0
}

func (this *roleAttributeMgr) List() []*model.RoleAttribute {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	ret := make([]*model.RoleAttribute, 0, len(this.attribute))
	for _, attribute := range this.attribute {
		ret = append(ret, attribute)
	}
	return ret
}
