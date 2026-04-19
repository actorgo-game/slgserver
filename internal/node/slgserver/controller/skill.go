package controller

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorSkill struct {
	cactor.Base
}

func NewActorSkill() *ActorSkill      { return &ActorSkill{} }
func (p *ActorSkill) AliasID() string { return "skill" }
func (p *ActorSkill) OnInit()         { clog.Info("[ActorSkill] initialized") }

func (p *ActorSkill) OnFindChild(m *cfacade.Message) (cfacade.IActor, bool) {
	a, err := p.Child().Create(m.TargetPath().ChildID, &SkillChild{})
	if err != nil {
		return nil, false
	}
	return a, true
}

type SkillChild struct{ userActor }

func (p *SkillChild) OnInit() {
	p.Remote().Register("list", p.list)
}

func (p *SkillChild) list(_ *protocol.SkillListReq) (*protocol.SkillListRsp, int32) {
	rsp := &protocol.SkillListRsp{List: make([]protocol.Skill, 0)}
	role, ec := p.MyRole()
	if ec != code.OK {
		return rsp, ec
	}
	skills, _ := mgr.SkillMgr.Get(role.RId)
	for _, sk := range skills {
		rsp.List = append(rsp.List, sk.ToProto().(protocol.Skill))
	}
	return rsp, code.OK
}
