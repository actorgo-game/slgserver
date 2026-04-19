package mgr

import (
	"sync"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/data/model"
)

type coalitionMgr struct {
	mutex  sync.RWMutex
	unions map[int]*model.Coalition // key: id
}

var UnionMgr = &coalitionMgr{
	unions: make(map[int]*model.Coalition),
}

func (this *coalitionMgr) Load() {
	c := coll(model.Coalition{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{"state": model.UnionRunning}))
	if err != nil {
		clog.Error("[UnionMgr] Load find err=%v", err)
		return
	}
	var rows []*model.Coalition
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[UnionMgr] Load decode err=%v", err)
		return
	}
	for _, v := range rows {
		this.unions[v.Id] = v
	}
}

func (this *coalitionMgr) Get(unionId int) (*model.Coalition, bool) {
	this.mutex.RLock()
	r, ok := this.unions[unionId]
	this.mutex.RUnlock()
	if ok {
		return r, true
	}

	c := coll(model.Coalition{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	m := &model.Coalition{}
	if err := c.FindOne(cx, withServer(bson.M{"id": unionId, "state": model.UnionRunning})).Decode(m); err != nil {
		return nil, false
	}
	this.mutex.Lock()
	this.unions[unionId] = m
	this.mutex.Unlock()
	return m, true
}

func (this *coalitionMgr) Create(name string, rid int) (*model.Coalition, bool) {
	c := coll(model.Coalition{}.CollectionName())
	if c == nil {
		return nil, false
	}
	m := &model.Coalition{
		ServerId:    serverID(),
		Id:          nextID(model.Coalition{}.CollectionName()),
		Name:        name,
		Ctime:       time.Now(),
		CreateId:    rid,
		Chairman:    rid,
		State:       model.UnionRunning,
		MemberArray: []int{rid},
	}
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, m); err != nil {
		clog.Error("[UnionMgr] Create insert err=%v", err)
		return nil, false
	}
	this.mutex.Lock()
	this.unions[m.Id] = m
	this.mutex.Unlock()
	return m, true
}

func (this *coalitionMgr) List() []*model.Coalition {
	r := make([]*model.Coalition, 0, len(this.unions))
	this.mutex.RLock()
	for _, c := range this.unions {
		r = append(r, c)
	}
	this.mutex.RUnlock()
	return r
}

func (this *coalitionMgr) Remove(unionId int) {
	this.mutex.Lock()
	delete(this.unions, unionId)
	this.mutex.Unlock()
}

// MainMembers 获取联盟的主要成员（chairman + viceChairman），用于 push 路由。
// 注入到 model.GetMainMembers。
func MainMembers(unionId int) []int {
	u, ok := UnionMgr.Get(unionId)
	if !ok {
		return nil
	}
	r := make([]int, 0, 2)
	if u.Chairman != 0 {
		r = append(r, u.Chairman)
	}
	if u.ViceChairman != 0 {
		r = append(r, u.ViceChairman)
	}
	return r
}

// UnionId 注入到 model.GetUnionId。
func UnionId(rid int) int { return RAttrMgr.UnionId(rid) }

// UnionName 注入到 model.GetUnionName。
func UnionName(unionId int) string {
	if u, ok := UnionMgr.Get(unionId); ok {
		return u.Name
	}
	return ""
}

// ParentId 注入到 model.GetParentId。
func ParentId(rid int) int {
	uid := RAttrMgr.UnionId(rid)
	if uid == 0 {
		return 0
	}
	if u, ok := UnionMgr.Get(uid); ok {
		return u.Chairman
	}
	return 0
}
