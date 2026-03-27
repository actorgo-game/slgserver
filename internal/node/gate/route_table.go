package gate

import (
	mynet "github.com/llr104/slgserver/internal/net"
)

type targetKind int

const (
	kindCenter targetKind = iota
	kindPlayer
	kindMap
	kindUnion
	kindChat
)

type afterHookFunc func(app interface{}, req *mynet.WsMsgReq, rsp *mynet.WsMsgRsp)

type routeEntry struct {
	method    string
	target    targetKind
	needLogin bool
	needRole  bool
	needRId   bool
	afterHook afterHookFunc
}

var routes = map[string][]routeEntry{
	"account": {
		{method: "login", target: kindCenter, afterHook: afterAccountLogin},
		{method: "reLogin", target: kindCenter, afterHook: afterAccountReLogin},
		{method: "logout", target: kindCenter, needLogin: true, afterHook: afterAccountLogout},
		{method: "serverList", target: kindCenter, needLogin: true},
	},
	"role": {
		{method: "enterServer", target: kindPlayer, needLogin: true, afterHook: afterEnterServer},
		{method: "create", target: kindPlayer, needLogin: true},
		{method: "roleList", target: kindPlayer, needLogin: true},
		{method: "myCity", target: kindPlayer, needRole: true},
		{method: "myRoleRes", target: kindPlayer, needRole: true},
		{method: "myRoleBuild", target: kindPlayer, needRole: true},
		{method: "myProperty", target: kindPlayer, needRole: true},
		{method: "upPosition", target: kindPlayer, needRole: true},
		{method: "posTagList", target: kindPlayer, needRole: true},
		{method: "opPosTag", target: kindPlayer, needRole: true},
	},
	"general": {
		{method: "myGenerals", target: kindPlayer, needLogin: true, needRole: true},
		{method: "drawGeneral", target: kindPlayer, needLogin: true, needRole: true},
		{method: "composeGeneral", target: kindPlayer, needLogin: true, needRole: true},
		{method: "addPrGeneral", target: kindPlayer, needLogin: true, needRole: true},
		{method: "convert", target: kindPlayer, needLogin: true, needRole: true},
		{method: "upSkill", target: kindPlayer, needLogin: true, needRole: true},
		{method: "downSkill", target: kindPlayer, needLogin: true, needRole: true},
		{method: "lvSkill", target: kindPlayer, needLogin: true, needRole: true},
	},
	"army": {
		{method: "myList", target: kindPlayer, needLogin: true, needRole: true},
		{method: "myOne", target: kindPlayer, needLogin: true, needRole: true},
		{method: "dispose", target: kindPlayer, needLogin: true, needRole: true},
		{method: "conscript", target: kindPlayer, needLogin: true, needRole: true},
		{method: "assign", target: kindPlayer, needLogin: true, needRole: true},
	},
	"city": {
		{method: "facilities", target: kindPlayer, needLogin: true, needRole: true},
		{method: "upFacility", target: kindPlayer, needLogin: true, needRole: true},
	},
	"interior": {
		{method: "collect", target: kindPlayer, needRole: true},
		{method: "openCollect", target: kindPlayer, needRole: true},
		{method: "transform", target: kindPlayer, needRole: true},
	},
	"skill": {
		{method: "list", target: kindPlayer, needLogin: true, needRole: true},
	},
	"war": {
		{method: "report", target: kindPlayer, needRole: true},
		{method: "read", target: kindPlayer, needRole: true},
	},
	"nationMap": {
		{method: "config", target: kindMap},
		{method: "scan", target: kindMap, needRole: true},
		{method: "scanBlock", target: kindMap, needRole: true},
		{method: "giveUp", target: kindMap, needRole: true},
		{method: "build", target: kindMap, needRole: true},
		{method: "upBuild", target: kindMap, needRole: true},
		{method: "delBuild", target: kindMap, needRole: true},
	},
	"union": {
		{method: "create", target: kindUnion, needLogin: true, needRole: true},
		{method: "list", target: kindUnion, needLogin: true, needRole: true},
		{method: "join", target: kindUnion, needLogin: true, needRole: true},
		{method: "verify", target: kindUnion, needLogin: true, needRole: true},
		{method: "member", target: kindUnion, needLogin: true, needRole: true},
		{method: "applyList", target: kindUnion, needLogin: true, needRole: true},
		{method: "exit", target: kindUnion, needLogin: true, needRole: true},
		{method: "dismiss", target: kindUnion, needLogin: true, needRole: true},
		{method: "notice", target: kindUnion, needLogin: true, needRole: true},
		{method: "modNotice", target: kindUnion, needLogin: true, needRole: true},
		{method: "kick", target: kindUnion, needLogin: true, needRole: true},
		{method: "appoint", target: kindUnion, needLogin: true, needRole: true},
		{method: "abdicate", target: kindUnion, needLogin: true, needRole: true},
		{method: "info", target: kindUnion, needLogin: true, needRole: true},
		{method: "log", target: kindUnion, needLogin: true, needRole: true},
	},
	"chat": {
		{method: "login", target: kindChat},
		{method: "logout", target: kindChat, needRId: true},
		{method: "chat", target: kindChat, needRId: true},
		{method: "history", target: kindChat, needRId: true},
		{method: "join", target: kindChat, needRId: true},
		{method: "exit", target: kindChat, needRId: true},
	},
}
