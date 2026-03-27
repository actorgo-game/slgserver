package net

import (
	"sync"

	"github.com/gorilla/websocket"

	clog "github.com/actorgo-game/actorgo/logger"
)

var ConnMgr = Mgr{}
var cid int64 = 0

type Mgr struct {
	cm sync.RWMutex
	um sync.RWMutex
	rm sync.RWMutex

	connCache map[int64]WSConn
	userCache map[int]WSConn
	roleCache map[int]WSConn
}

func (mgr *Mgr) NewConn(wsSocket *websocket.Conn, needSecret bool) *ServerConn {
	mgr.cm.Lock()
	defer mgr.cm.Unlock()

	cid++
	if mgr.connCache == nil {
		mgr.connCache = make(map[int64]WSConn)
	}
	if mgr.userCache == nil {
		mgr.userCache = make(map[int]WSConn)
	}
	if mgr.roleCache == nil {
		mgr.roleCache = make(map[int]WSConn)
	}

	c := NewServerConn(wsSocket, needSecret)
	c.SetProperty("cid", cid)
	mgr.connCache[cid] = c
	return c
}

func (mgr *Mgr) UserLogin(conn WSConn, session string, uid int) {
	mgr.um.Lock()
	defer mgr.um.Unlock()

	oldConn, ok := mgr.userCache[uid]
	if ok {
		if conn != oldConn {
			clog.Info("rob login: uid=%d, old=%s, new=%s", uid, oldConn.Addr(), conn.Addr())
			oldConn.Push("robLogin", nil)
		}
	}
	mgr.userCache[uid] = conn
	conn.SetProperty("session", session)
	conn.SetProperty("uid", uid)
}

func (mgr *Mgr) UserLogout(conn WSConn) {
	mgr.removeUser(conn)
}

func (mgr *Mgr) removeUser(conn WSConn) {
	mgr.um.Lock()
	uid, err := conn.GetProperty("uid")
	if err == nil {
		id := uid.(int)
		c, ok := mgr.userCache[id]
		if ok && c == conn {
			delete(mgr.userCache, id)
		}
	}
	mgr.um.Unlock()

	mgr.rm.Lock()
	rid, err := conn.GetProperty("rid")
	if err == nil {
		id := rid.(int)
		c, ok := mgr.roleCache[id]
		if ok && c == conn {
			delete(mgr.roleCache, id)
		}
	}
	mgr.rm.Unlock()

	conn.RemoveProperty("session")
	conn.RemoveProperty("uid")
	conn.RemoveProperty("role")
	conn.RemoveProperty("rid")
}

func (mgr *Mgr) RoleEnter(conn WSConn, rid int) {
	mgr.rm.Lock()
	defer mgr.rm.Unlock()
	conn.SetProperty("rid", rid)
	mgr.roleCache[rid] = conn
}

func (mgr *Mgr) RemoveConn(conn WSConn) {
	mgr.cm.Lock()
	cidVal, err := conn.GetProperty("cid")
	if err == nil {
		delete(mgr.connCache, cidVal.(int64))
		conn.RemoveProperty("cid")
	}
	mgr.cm.Unlock()

	mgr.removeUser(conn)
}

func (mgr *Mgr) PushByRoleId(rid int, msgName string, data interface{}) bool {
	if rid <= 0 {
		return false
	}
	mgr.rm.RLock()
	defer mgr.rm.RUnlock()
	conn, ok := mgr.roleCache[rid]
	if ok {
		conn.Push(msgName, data)
		return true
	}
	return false
}

func (mgr *Mgr) Count() int {
	mgr.cm.RLock()
	defer mgr.cm.RUnlock()
	return len(mgr.connCache)
}
