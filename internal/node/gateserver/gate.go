package gateserver

import (
	"github.com/actorgo-game/actorgo"
	cfacade "github.com/actorgo-game/actorgo/facade"
	cconnector "github.com/actorgo-game/actorgo/net/connector"
	"github.com/actorgo-game/actorgo/net/parser/pomelo"
)

func Run(profileFilePath, nodeID string) {
	app := actorgo.Configure(profileFilePath, nodeID, true, actorgo.Cluster)

	// 设置网络数据包解析器
	netParser := buildPomeloParser(app)
	//netParser := buildSimpleParser(app)
	app.SetNetParser(netParser)

	app.Startup()
}

func buildPomeloParser(app *actorgo.AppBuilder) cfacade.INetParser {
	// 使用pomelo网络数据包解析器
	agentActor := pomelo.NewActor("user")
	//创建一个tcp监听，用于client/robot压测机器人连接网关tcp
	//agentActor.AddConnector(cconnector.NewTCP(":8004"))
	//再创建一个websocket监听，用于h5客户端建立连接
	agentActor.AddConnector(cconnector.NewWS(app.Address()))
	//当有新连接创建Agent时，启动一个自定义(ActorAgent)的子actor
	agentActor.SetOnNewAgent(func(newAgent *pomelo.Agent) {
		childActor := &ActorAgent{}
		newAgent.AddOnClose(childActor.onSessionClose)
		agentActor.Child().Create(newAgent.SID(), childActor) // actorID == sid
	})

	// 设置数据路由函数
	agentActor.SetOnDataRoute(onPomeloDataRoute)

	return agentActor
}
