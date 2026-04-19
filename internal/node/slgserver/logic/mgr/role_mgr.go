package mgr

import (
	"sync"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/data/model"
)

// RoleNickName 与原 mgr.RoleNickName 同名同语义，注入到 model.GetRoleNickName。
func RoleNickName(rid int) string {
	if r, ok := RMgr.Get(rid); ok {
		return r.NickName
	}
	return ""
}

type roleMgr struct {
	mutex sync.RWMutex
	roles map[int]*model.Role // key: rid
}

var RMgr = &roleMgr{
	roles: make(map[int]*model.Role),
}

// Load 拉取所有角色到内存。
func (this *roleMgr) Load() {
	c := coll(model.Role{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[RMgr] Load find err=%v", err)
		return
	}
	var rows []*model.Role
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[RMgr] Load decode err=%v", err)
		return
	}
	this.mutex.Lock()
	for _, r := range rows {
		this.roles[r.RId] = r
	}
	this.mutex.Unlock()
}

func (this *roleMgr) Get(rid int) (*model.Role, bool) {
	this.mutex.RLock()
	r, ok := this.roles[rid]
	this.mutex.RUnlock()
	if ok {
		return r, true
	}

	c := coll(model.Role{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()

	m := &model.Role{}
	if err := c.FindOne(cx, withServer(bson.M{"rid": rid})).Decode(m); err != nil {
		clog.Warn("[RMgr] Get rid=%d err=%v", rid, err)
		return nil, false
	}

	this.mutex.Lock()
	this.roles[rid] = m
	this.mutex.Unlock()
	return m, true
}

// GetByUId 用于 enterServer 入口：以 uid 找角色。
func (this *roleMgr) GetByUId(uid int) (*model.Role, bool) {
	this.mutex.RLock()
	for _, r := range this.roles {
		if r.UId == uid {
			this.mutex.RUnlock()
			return r, true
		}
	}
	this.mutex.RUnlock()

	c := coll(model.Role{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()

	m := &model.Role{}
	if err := c.FindOne(cx, withServer(bson.M{"uid": uid})).Decode(m); err != nil {
		return nil, false
	}
	this.mutex.Lock()
	this.roles[m.RId] = m
	this.mutex.Unlock()
	return m, true
}

// ListByUId 列出该 uid 下所有角色（多角色时使用）。
func (this *roleMgr) ListByUId(uid int) []*model.Role {
	c := coll(model.Role{}.CollectionName())
	if c == nil {
		return nil
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{"uid": uid}))
	if err != nil {
		clog.Warn("[RMgr] ListByUId err=%v", err)
		return nil
	}
	var roles []*model.Role
	if err := cursor.All(cx, &roles); err != nil {
		return nil
	}
	this.mutex.Lock()
	for _, r := range roles {
		this.roles[r.RId] = r
	}
	this.mutex.Unlock()
	return roles
}

// Add 把角色加入缓存（创建新角色时调用）。
func (this *roleMgr) Add(r *model.Role) {
	this.mutex.Lock()
	this.roles[r.RId] = r
	this.mutex.Unlock()
}

// Insert 写入新角色到 mongo，自增 rid，并加入缓存。
func (this *roleMgr) Insert(r *model.Role) error {
	c := coll(model.Role{}.CollectionName())
	if c == nil {
		return ErrDBNotReady
	}
	if r.RId == 0 {
		r.RId = nextID(model.Role{}.CollectionName())
	}
	r.ServerId = serverID()
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, r); err != nil {
		return err
	}
	this.Add(r)
	return nil
}
