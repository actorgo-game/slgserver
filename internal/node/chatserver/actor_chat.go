package chatserver

import (
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/protocol"
)

const maxWorldHistory = 100

// Internal request types for Remote handlers that need caller identification.

type chatSendReq struct {
	RId  int    `json:"rid"`
	Type int8   `json:"type"`
	Msg  string `json:"msg"`
}

type chatSendRsp struct {
	Msg protocol.ChatMsg `json:"msg"`
}

type chatHistReq struct {
	RId  int  `json:"rid"`
	Type int8 `json:"type"`
}

type chatJoinReq struct {
	RId int `json:"rid"`
	Id  int `json:"id"`
}

type chatExitReq struct {
	RId int `json:"rid"`
	Id  int `json:"id"`
}

type chatUser struct {
	RId      int
	NickName string
}

type ActorChatRoom struct {
	cactor.Base

	worldMsgs  []protocol.ChatMsg
	unionMsgs  map[int][]protocol.ChatMsg
	users      map[int]*chatUser
	ridToUnion map[int]int
}

func NewActorChatRoom() *ActorChatRoom {
	return &ActorChatRoom{
		worldMsgs:  make([]protocol.ChatMsg, 0, maxWorldHistory),
		unionMsgs:  make(map[int][]protocol.ChatMsg),
		users:      make(map[int]*chatUser),
		ridToUnion: make(map[int]int),
	}
}

func (p *ActorChatRoom) AliasID() string {
	return "room"
}

func (p *ActorChatRoom) OnInit() {
	p.Remote().Register("login", p.login)
	p.Remote().Register("logout", p.logout)
	p.Remote().Register("chat", p.chat)
	p.Remote().Register("history", p.history)
	p.Remote().Register("join", p.join)
	p.Remote().Register("exit", p.exit)

	clog.Info("[ActorChatRoom] initialized")
}

func (p *ActorChatRoom) login(req *protocol.ChatLoginReq) (*protocol.ChatLoginRsp, int32) {
	p.users[req.RId] = &chatUser{
		RId:      req.RId,
		NickName: req.NickName,
	}
	return &protocol.ChatLoginRsp{RId: req.RId, NickName: req.NickName}, code.OK
}

func (p *ActorChatRoom) logout(req *protocol.ChatLogoutReq) (*protocol.ChatLogoutRsp, int32) {
	delete(p.users, req.RId)
	return &protocol.ChatLogoutRsp{RId: req.RId}, code.OK
}

func (p *ActorChatRoom) chat(req *chatSendReq) (*chatSendRsp, int32) {
	rid := req.RId
	user := p.users[rid]
	nickName := ""
	if user != nil {
		nickName = user.NickName
	}

	msg := protocol.ChatMsg{
		RId:      rid,
		NickName: nickName,
		Type:     req.Type,
		Msg:      req.Msg,
		Time:     time.Now().Unix(),
	}

	if req.Type == 0 {
		p.worldMsgs = append(p.worldMsgs, msg)
		if len(p.worldMsgs) > maxWorldHistory {
			p.worldMsgs = p.worldMsgs[len(p.worldMsgs)-maxWorldHistory:]
		}
	} else if req.Type == 1 {
		unionId, ok := p.ridToUnion[rid]
		if ok {
			p.unionMsgs[unionId] = append(p.unionMsgs[unionId], msg)
			if len(p.unionMsgs[unionId]) > maxWorldHistory {
				p.unionMsgs[unionId] = p.unionMsgs[unionId][len(p.unionMsgs[unionId])-maxWorldHistory:]
			}
		}
	}

	return &chatSendRsp{Msg: msg}, code.OK
}

func (p *ActorChatRoom) history(req *chatHistReq) (*protocol.ChatHistoryRsp, int32) {
	rsp := &protocol.ChatHistoryRsp{Type: req.Type, Msgs: make([]protocol.ChatMsg, 0)}

	if req.Type == 0 {
		rsp.Msgs = p.worldMsgs
	} else if req.Type == 1 {
		if unionId, ok := p.ridToUnion[req.RId]; ok {
			rsp.Msgs = p.unionMsgs[unionId]
		}
	}

	return rsp, code.OK
}

func (p *ActorChatRoom) join(req *chatJoinReq) int32 {
	p.ridToUnion[req.RId] = req.Id
	return code.OK
}

func (p *ActorChatRoom) exit(req *chatExitReq) int32 {
	delete(p.ridToUnion, req.RId)
	return code.OK
}
