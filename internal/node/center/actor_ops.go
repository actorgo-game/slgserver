package center

import (
	cactor "github.com/actorgo-game/actorgo/net/actor"
)

type ActorOps struct {
	cactor.Base
}

func (p *ActorOps) AliasID() string {
	return "ops"
}

func (p *ActorOps) OnInit() {
	p.Remote().Register("ping", p.ping)
}

type boolRsp struct {
	Value bool `json:"value"`
}

func (p *ActorOps) ping() (*boolRsp, int32) {
	return &boolRsp{Value: true}, 0
}
