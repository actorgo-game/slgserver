package web

import (
	"net/http"

	"github.com/actorgo-game/actorgo"
	cgin "github.com/actorgo-game/actorgo/components/gin"
	cfacade "github.com/actorgo-game/actorgo/facade"
	"github.com/gin-gonic/gin"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/rpc"
)

var appInstance cfacade.IApplication

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	appInstance = app

	httpServer := cgin.NewHttp("http_server", app.Address())
	gin.SetMode(gin.DebugMode)
	httpServer.Use(cgin.Cors())
	httpServer.Use(cgin.RecoveryWithZap(true))

	httpServer.POST("/account/register", handleRegister)
	httpServer.POST("/account/login", handleLogin)
	httpServer.POST("/account/changepwd", handleChangePwd)
	httpServer.POST("/account/forgetpwd", handleForgetPwd)
	httpServer.POST("/account/resetpwd", handleResetPwd)
	httpServer.GET("/server/list", handleServerList)

	app.Register(httpServer)
	app.Startup()
}

func handleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": code.InvalidParam, "msg": "invalid params"})
		return
	}

	errCode := rpc.Register(appInstance, req.Username, req.Password, c.ClientIP())
	if code.IsFail(errCode) {
		c.JSON(http.StatusOK, gin.H{"code": errCode, "msg": "register failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code.OK, "msg": "ok", "data": gin.H{"username": req.Username}})
}

func handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": code.InvalidParam, "msg": "invalid params"})
		return
	}

	rsp, errCode := rpc.Login(appInstance, req.Username, req.Password)
	if code.IsFail(errCode) || rsp == nil {
		c.JSON(http.StatusOK, gin.H{"code": errCode, "msg": "login failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code.OK,
		"msg":  "ok",
		"data": gin.H{
			"uid":      rsp.UId,
			"username": rsp.Username,
			"token":    rsp.Session,
		},
	})
}

func handleChangePwd(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		OldPwd   string `json:"oldPwd"`
		NewPwd   string `json:"newPwd"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": code.InvalidParam, "msg": "invalid params"})
		return
	}

	errCode := rpc.ChangePwd(appInstance, req.Username, req.OldPwd, req.NewPwd)
	if code.IsFail(errCode) {
		c.JSON(http.StatusOK, gin.H{"code": errCode, "msg": "change password failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code.OK, "msg": "ok"})
}

func handleForgetPwd(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": code.InvalidParam, "msg": "invalid params"})
		return
	}

	username, errCode := rpc.ForgetPwd(appInstance, req.Username)
	if code.IsFail(errCode) {
		c.JSON(http.StatusOK, gin.H{"code": errCode, "msg": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code.OK, "msg": "ok", "data": gin.H{"username": username}})
}

func handleResetPwd(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		NewPwd   string `json:"newPwd"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": code.InvalidParam, "msg": "invalid params"})
		return
	}

	errCode := rpc.ResetPwd(appInstance, req.Username, req.NewPwd)
	if code.IsFail(errCode) {
		c.JSON(http.StatusOK, gin.H{"code": errCode, "msg": "reset password failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code.OK, "msg": "ok"})
}

func handleServerList(c *gin.Context) {
	servers := []gin.H{
		{"id": 1, "name": "SLG Server 1", "address": "ws://127.0.0.1:9010"},
	}
	c.JSON(http.StatusOK, gin.H{"code": code.OK, "msg": "ok", "data": servers})
}
