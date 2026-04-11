package httpserver

import (
	"github.com/actorgo-game/actorgo"
	cgin "github.com/actorgo-game/actorgo/components/gin"
	cfacade "github.com/actorgo-game/actorgo/facade"
	"github.com/gin-gonic/gin"
)

var appInstance cfacade.IApplication

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	appInstance = app

	httpServer := cgin.NewHttp("http_server", app.Address())
	gin.SetMode(gin.DebugMode)
	httpServer.Use(cgin.Cors())
	httpServer.Use(cgin.RecoveryWithZap(true))
	//注册 controller
	httpServer.Register(new(Controller))

	app.Register(httpServer)
	app.Startup()
}
