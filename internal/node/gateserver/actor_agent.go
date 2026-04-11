package gateserver

import (
	cstring "github.com/actorgo-game/actorgo/extend/string"
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/actorgo-game/actorgo/net/parser/pomelo"
	cproto "github.com/actorgo-game/actorgo/net/proto"
	"github.com/llr104/slgserver/internal/code"
	pb "github.com/llr104/slgserver/internal/protocol"
	"github.com/llr104/slgserver/internal/util"
)

var (
	duplicateLoginCode []byte
)

type (
	// ActorAgent 每个网络连接对应一个ActorAgent
	ActorAgent struct {
		cactor.Base
	}
)

func (p *ActorAgent) OnInit() {
	duplicateLoginCode, _ = p.App().Serializer().Marshal(&cproto.I32{
		Value: code.PlayerDuplicateLogin,
	})

	p.Local().Register("login", p.login)
	p.Remote().Register("setSession", p.setSession)
}

func (p *ActorAgent) setSession(req *pb.StringKeyValue) {
	if req.Key == "" {
		return
	}

	if agent, ok := pomelo.GetAgent(p.ActorID(), 0); ok {
		agent.Session().Set(req.Key, req.Value)
	}
}

// login 用户登录，验证帐号 (*pb.LoginResponse, int32)
func (p *ActorAgent) login(session *cproto.Session, req *pb.LoginRequest) {
	agent, found := pomelo.GetAgent(p.ActorID(), session.Uid)
	if !found {
		return
	}

	// 验证token
	userToken, errCode := p.validateToken(req.Token)
	if code.IsFail(errCode) {
		agent.Response(session, errCode)
		return
	}

	// 根据token带来的sdk参数，从中心节点获取uid
	uid := int64(0)
	/*uid, errCode := rpcCenter.GetUID(p.App(), userToken.OpenID)
	if uid == 0 || code.IsFail(errCode) {
		agent.ResponseCode(session, code.AccountBindFail, true)
		return
	}*/

	oldAgent, err := pomelo.Bind(session.Sid, uid)
	if err != nil {
		agent.ResponseCode(session, code.AccountBindFail, true)
		clog.Warn(err.Error())
		return
	}

	// 挤掉之前的agent
	if oldAgent != nil {
		oldAgent.Kick(duplicateLoginCode, true)
	}

	p.checkGateSession(uid)

	agent.Session().Set(util.ServerID, cstring.ToString(req.ServerId))
	agent.Session().Set(util.PID, cstring.ToString(userToken.PID))
	agent.Session().Set(util.OpenID, userToken.OpenID)

	response := &pb.LoginResponse{
		Uid:    uid,
		Pid:    userToken.PID,
		OpenId: userToken.OpenID,
	}

	agent.Response(session, response)
}

func (p *ActorAgent) validateToken(base64Token string) (*util.Token, int32) {
	userToken, ok := util.DecodeToken(base64Token)
	if !ok {
		return nil, code.AccountTokenValidateFail
	}

	return userToken, code.OK
}

func (p *ActorAgent) checkGateSession(uid cfacade.UID) {
	rsp := &cproto.PomeloKick{
		Uid:    uid,
		Reason: duplicateLoginCode,
	}

	// 遍历其他网关节点，挤掉旧的agent
	members := p.App().Discovery().ListByType(p.App().NodeType(), p.App().NodeID())
	for _, member := range members {
		// user是gate.go里自定义的agentActorID
		actorPath := cfacade.NewPath(member.GetNodeID(), "user")
		p.Call(actorPath, pomelo.KickFuncName, rsp)
	}
}

// onSessionClose  当agent断开时，关闭对应的ActorAgent
func (p *ActorAgent) onSessionClose(agent *pomelo.Agent) {
	session := agent.Session()
	serverId := session.GetString(util.ServerID)
	if serverId == "" {
		return
	}

	// 通知game节点关闭session
	childId := cstring.ToString(session.Uid)
	if childId != "" {
		targetPath := cfacade.NewChildPath(serverId, "player", childId)
		p.Call(targetPath, "sessionClose", nil)
	}

	// 自己退出
	p.Exit()
	clog.Info("sessionClose path = %s", p.Path())
}
