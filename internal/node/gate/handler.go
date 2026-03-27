package gate

import (
	"strconv"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	"github.com/llr104/slgserver/internal/code"
	mynet "github.com/llr104/slgserver/internal/net"
	"github.com/llr104/slgserver/internal/protocol"
)

const sourcePath = ".gate"

func nodeID(app cfacade.IApplication, nodeType string) string {
	list := app.Discovery().ListByType(nodeType)
	if len(list) > 0 {
		return list[0].GetNodeID()
	}
	return ""
}

func buildRouter(r *mynet.Router, app cfacade.IApplication) {
	for prefix, entries := range routes {
		g := r.Group(prefix).Use(mynet.ElapsedTime(), mynet.Log())

		needGroupLogin := allNeed(entries, func(e routeEntry) bool { return e.needLogin })
		needGroupRole := allNeed(entries, func(e routeEntry) bool { return e.needRole })
		needGroupRId := allNeed(entries, func(e routeEntry) bool { return e.needRId })

		if needGroupLogin {
			g.Use(mynet.CheckLogin())
		}
		if needGroupRole {
			g.Use(mynet.CheckRole())
		}
		if needGroupRId {
			g.Use(mynet.CheckRId())
		}

		for _, e := range entries {
			entry := e
			var perRoute []mynet.MiddlewareFunc
			if entry.needLogin && !needGroupLogin {
				perRoute = append(perRoute, mynet.CheckLogin())
			}
			if entry.needRole && !needGroupRole {
				perRoute = append(perRoute, mynet.CheckRole())
			}
			if entry.needRId && !needGroupRId {
				perRoute = append(perRoute, mynet.CheckRId())
			}

			handler := makeHandler(app, entry)
			g.AddRouter(entry.method, handler, perRoute...)
		}
	}
}

func allNeed(entries []routeEntry, pred func(routeEntry) bool) bool {
	for _, e := range entries {
		if !pred(e) {
			return false
		}
	}
	return len(entries) > 0
}

func makeHandler(app cfacade.IApplication, entry routeEntry) mynet.HandlerFunc {
	return func(req *mynet.WsMsgReq, rsp *mynet.WsMsgRsp) {
		forwardToActor(app, entry.target, entry.method, req, rsp)
		if entry.afterHook != nil && rsp.Body.Code == int(code.OK) {
			entry.afterHook(app, req, rsp)
		}
	}
}

func resolveTarget(app cfacade.IApplication, kind targetKind, req *mynet.WsMsgReq) (string, bool) {
	switch kind {
	case kindCenter:
		nid := nodeID(app, "center")
		if nid == "" {
			return "", false
		}
		return nid + ".account", true

	case kindPlayer:
		uidVal, err := req.Conn.GetProperty("uid")
		if err != nil {
			return "", false
		}
		uid, ok := uidVal.(int)
		if !ok || uid == 0 {
			return "", false
		}
		nid := nodeID(app, "game")
		if nid == "" {
			return "", false
		}
		return cfacade.NewChildPath(nid, "player", strconv.Itoa(uid)), true

	case kindMap:
		nid := nodeID(app, "game")
		if nid == "" {
			return "", false
		}
		return nid + ".map", true

	case kindUnion:
		nid := nodeID(app, "game")
		if nid == "" {
			return "", false
		}
		return nid + ".union", true

	case kindChat:
		nid := nodeID(app, "chat")
		if nid == "" {
			return "", false
		}
		return nid + ".room", true
	}
	return "", false
}

func forwardToActor(app cfacade.IApplication, kind targetKind, method string,
	req *mynet.WsMsgReq, rsp *mynet.WsMsgRsp) {

	targetPath, ok := resolveTarget(app, kind, req)
	if !ok {
		if kind == kindPlayer {
			rsp.Body.Code = int(code.SessionInvalid)
		} else {
			rsp.Body.Code = int(code.Error)
		}
		clog.Warn("[gate] resolve target failed: kind=%d, method=%s", kind, method)
		return
	}

	rspMsg := make(map[string]interface{})
	errCode := app.ActorSystem().CallWait(sourcePath, targetPath, method, req.Body.Msg, &rspMsg)

	rsp.Body.Code = int(errCode)
	if errCode == code.OK {
		rsp.Body.Msg = rspMsg
	}
}

// ---------------------------------------------------------------------------
// After-hooks for routes that need post-processing
// ---------------------------------------------------------------------------

func afterAccountLogin(_ interface{}, req *mynet.WsMsgReq, rsp *mynet.WsMsgRsp) {
	data, ok := rsp.Body.Msg.(map[string]interface{})
	if !ok {
		return
	}
	uidFloat, _ := data["uid"].(float64)
	uid := int(uidFloat)
	session, _ := data["session"].(string)
	if uid > 0 && session != "" {
		mynet.ConnMgr.UserLogin(req.Conn, session, uid)
	}
}

func afterAccountReLogin(_ interface{}, req *mynet.WsMsgReq, rsp *mynet.WsMsgRsp) {
	data, ok := rsp.Body.Msg.(map[string]interface{})
	if !ok {
		return
	}
	session, _ := data["session"].(string)
	uidVal, err := req.Conn.GetProperty("uid")
	if err != nil {
		return
	}
	uid, ok := uidVal.(int)
	if ok && uid != 0 && session != "" {
		mynet.ConnMgr.UserLogin(req.Conn, session, uid)
	}
}

func afterAccountLogout(_ interface{}, req *mynet.WsMsgReq, _ *mynet.WsMsgRsp) {
	mynet.ConnMgr.UserLogout(req.Conn)
}

func afterEnterServer(appRaw interface{}, req *mynet.WsMsgReq, rsp *mynet.WsMsgRsp) {
	data, ok := rsp.Body.Msg.(map[string]interface{})
	if !ok {
		return
	}

	roleData, _ := data["role"].(map[string]interface{})
	if roleData == nil {
		return
	}

	ridFloat, _ := roleData["rid"].(float64)
	rid := int(ridFloat)
	if rid <= 0 {
		return
	}

	req.Conn.SetProperty("role", roleData)
	mynet.ConnMgr.RoleEnter(req.Conn, rid)

	app, ok := appRaw.(cfacade.IApplication)
	if !ok {
		return
	}

	chatNid := nodeID(app, "chat")
	if chatNid == "" {
		return
	}

	nickName, _ := roleData["nickName"].(string)
	token, _ := data["token"].(string)
	chatReq := &protocol.ChatLoginReq{
		RId:      rid,
		NickName: nickName,
		Token:    token,
	}
	chatRsp := make(map[string]interface{})
	errCode := app.ActorSystem().CallWait(sourcePath, chatNid+".room", "login", chatReq, &chatRsp)
	if errCode != code.OK {
		clog.Warn("[gate] auto chat login failed: rid=%d, errCode=%d", rid, errCode)
	}
}

// ---------------------------------------------------------------------------
// Connection close handler
// ---------------------------------------------------------------------------

func makeOnConnClose(app cfacade.IApplication) func(mynet.WSConn) {
	return func(conn mynet.WSConn) {
		uidVal, err := conn.GetProperty("uid")
		if err != nil {
			return
		}
		uid, ok := uidVal.(int)
		if !ok || uid == 0 {
			return
		}

		gameNid := nodeID(app, "game")
		if gameNid == "" {
			return
		}

		childID := strconv.Itoa(uid)
		targetPath := cfacade.NewChildPath(gameNid, "player", childID)
		errCode := app.ActorSystem().CallWait(sourcePath, targetPath, "sessionClose", nil, nil)
		if errCode != 0 {
			clog.Warn("[gate] notify sessionClose failed: uid=%d, errCode=%d", uid, errCode)
		}
	}
}
