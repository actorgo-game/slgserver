// Package controller 是 slgserver 节点的业务消息接入层（actorgo 版）。
//
// 与原 slgserver 的 controller 结构对应：
//
//	role.go        -> ActorRole / RoleChild
//	city.go        -> ActorCity / CityChild
//	general.go     -> ActorGeneral / GeneralChild
//	skill.go       -> ActorSkill / SkillChild
//	army.go        -> ActorArmy / ArmyChild
//	nation_map.go  -> ActorNationMap / NationMapChild
//	war.go         -> ActorWar / WarChild
//	coalition.go   -> ActorCoalition / CoalitionChild
//	interior.go    -> ActorInterior / InteriorChild
//
// 消息路由由 gateserver/route.go 完成：
//
//	cfacade.NewChildPath(serverId, route.HandleName(), uid)
//
// 因此每个 controller 的 AliasID 与原项目 group 名一致（"role"/"city"...），
// 父 actor 的 OnFindChild 按 uid 创建一个子 actor 处理该用户的所有请求，
// 子 actor 用 ActorID() 拿到 uid。
package controller

import (
	"strconv"

	cactor "github.com/actorgo-game/actorgo/net/actor"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/logic/mgr"
)

// userActor 所有 controller 子 actor 的公共字段：缓存 uid 与 role。
type userActor struct {
	cactor.Base
	uid int
}

func (p *userActor) UId() int {
	if p.uid == 0 {
		if v, err := strconv.Atoi(p.ActorID()); err == nil {
			p.uid = v
		}
	}
	return p.uid
}

// MyRole 拉取当前用户的角色对象。
func (p *userActor) MyRole() (*model.Role, int32) {
	uid := p.UId()
	if uid == 0 {
		return nil, code.PlayerIDError
	}
	r, ok := mgr.RMgr.GetByUId(uid)
	if !ok {
		return nil, code.RoleNotExist
	}
	return r, code.OK
}
