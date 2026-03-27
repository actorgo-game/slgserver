package player

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/llr104/slgserver/internal/node/game/module/online"
)

type ActorPlayers struct {
	cactor.Base
}

func (p *ActorPlayers) AliasID() string {
	return "player"
}

func (p *ActorPlayers) OnInit() {
	clog.Info("[ActorPlayers] init, nodeID=%s", p.App().NodeID())
}

func (p *ActorPlayers) OnFindChild(msg *cfacade.Message) (cfacade.IActor, bool) {
	childID := msg.TargetPath().ChildID
	childActor, err := p.Child().Create(childID, NewActorPlayer())
	if err != nil {
		clog.Warn("[ActorPlayers] create child actor fail: childID=%s, err=%v", childID, err)
		return nil, false
	}
	return childActor, true
}

func (p *ActorPlayers) OnStop() {
	clog.Info("[ActorPlayers] stop, onlineCount=%d", online.Count())
}
