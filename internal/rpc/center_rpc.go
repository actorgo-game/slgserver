package rpc

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/protocol"
)

const (
	centerType   = "center"
	accountActor = ".account"
	sourcePath   = ".system"
)

func GetCenterNodeID(app cfacade.IApplication) string {
	list := app.Discovery().ListByType(centerType)
	if len(list) > 0 {
		return list[0].GetNodeID()
	}
	return ""
}

func centerTarget(app cfacade.IApplication) string {
	return GetCenterNodeID(app) + accountActor
}

func Register(app cfacade.IApplication, username, password, ip string) int32 {
	req := &protocol.RegisterReq{
		Username: username,
		Password: password,
	}
	rsp := &protocol.RegisterRsp{}
	errCode := app.ActorSystem().CallWait(sourcePath, centerTarget(app), "register", req, rsp)
	if code.IsFail(errCode) {
		clog.Warn("[rpc.Register] username=%s errCode=%d", username, errCode)
	}
	return errCode
}

func Login(app cfacade.IApplication, username, password string) (*protocol.LoginRsp, int32) {
	req := &protocol.LoginReq{
		Username: username,
		Password: password,
	}
	rsp := &protocol.LoginRsp{}
	errCode := app.ActorSystem().CallWait(sourcePath, centerTarget(app), "login", req, rsp)
	if code.IsFail(errCode) {
		clog.Warn("[rpc.Login] username=%s errCode=%d", username, errCode)
		return nil, errCode
	}
	return rsp, code.OK
}

func ChangePwd(app cfacade.IApplication, username, oldPwd, newPwd string) int32 {
	type req struct {
		Username string `json:"username"`
		OldPwd   string `json:"oldPwd"`
		NewPwd   string `json:"newPwd"`
	}
	return app.ActorSystem().CallWait(sourcePath, centerTarget(app), "changePwd", &req{
		Username: username, OldPwd: oldPwd, NewPwd: newPwd,
	}, nil)
}

func ForgetPwd(app cfacade.IApplication, username string) (string, int32) {
	type req struct {
		Username string `json:"username"`
	}
	type rsp struct {
		Username string `json:"username"`
	}
	r := &rsp{}
	errCode := app.ActorSystem().CallWait(sourcePath, centerTarget(app), "forgetPwd", &req{Username: username}, r)
	if code.IsFail(errCode) {
		return "", errCode
	}
	return r.Username, code.OK
}

func ResetPwd(app cfacade.IApplication, username, newPwd string) int32 {
	type req struct {
		Username string `json:"username"`
		NewPwd   string `json:"newPwd"`
	}
	return app.ActorSystem().CallWait(sourcePath, centerTarget(app), "resetPwd", &req{
		Username: username, NewPwd: newPwd,
	}, nil)
}

func PingCenter(app cfacade.IApplication) bool {
	nodeID := GetCenterNodeID(app)
	if nodeID == "" {
		return false
	}
	type boolRsp struct {
		Value bool `json:"value"`
	}
	rsp := &boolRsp{}
	errCode := app.ActorSystem().CallWait(sourcePath, nodeID+".ops", "ping", nil, rsp)
	if code.IsFail(errCode) {
		return false
	}
	return rsp.Value
}
