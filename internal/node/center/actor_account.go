package center

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/go-redis/redis/v8"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/component"
	"github.com/llr104/slgserver/internal/data/entry"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/protocol"
)

type ActorAccount struct {
	cactor.Base
	accountEntry *entry.AccountEntry
	redisClient  *redis.Client
	uidCounter   int64
}

func (p *ActorAccount) AliasID() string {
	return "account"
}

func (p *ActorAccount) OnInit() {
	mongoComp := p.App().Find(component.MongoComponentName).(*component.MongoComponent)
	db := mongoComp.GetDb("slg_db")
	if db == nil {
		clog.Error("[ActorAccount] slg_db not found")
		return
	}
	p.accountEntry = entry.NewAccountEntry(db.Collection("accounts"))

	redisComp := p.App().Find(component.RedisComponentName).(*component.RedisComponent)
	p.redisClient = redisComp.Client()

	p.Remote().Register("register", p.register)
	p.Remote().Register("login", p.login)
	p.Remote().Register("reLogin", p.reLogin)
	p.Remote().Register("logout", p.logout)
	p.Remote().Register("serverList", p.serverList)
	p.Remote().Register("changePwd", p.changePwd)
	p.Remote().Register("forgetPwd", p.forgetPwd)
	p.Remote().Register("resetPwd", p.resetPwd)
}

func (p *ActorAccount) register(req *protocol.RegisterReq) int32 {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return code.InvalidParam
	}
	if len(req.Username) < 3 || len(req.Username) > 18 {
		return code.InvalidParam
	}

	existing, _ := p.accountEntry.FindByUsername(req.Username)
	if existing != nil {
		return code.UserExist
	}

	p.uidCounter++
	acc := &model.Account{
		AccountId:  p.uidCounter,
		Username:   req.Username,
		Password:   req.Password,
		CreateTime: time.Now(),
	}

	if err := p.accountEntry.Insert(acc); err != nil {
		clog.Warn("[ActorAccount.register] insert fail: %v", err)
		return code.DBError
	}

	return code.OK
}

func (p *ActorAccount) login(req *protocol.LoginReq) (*protocol.LoginRsp, int32) {
	acc, err := p.accountEntry.FindByUsername(req.Username)
	if err != nil || acc == nil {
		return nil, code.UserNotExist
	}

	if acc.Password != req.Password {
		return nil, code.PwdIncorrect
	}

	token := p.generateToken(acc.AccountId, req.Username)

	ctx := context.Background()
	p.redisClient.Set(ctx, fmt.Sprintf("token:%s", token), acc.AccountId, 24*time.Hour)
	p.redisClient.Set(ctx, fmt.Sprintf("session:%s", token), acc.AccountId, 24*time.Hour)

	return &protocol.LoginRsp{
		UId:      int(acc.AccountId),
		Username: acc.Username,
		Session:  token,
	}, code.OK
}

func (p *ActorAccount) reLogin(req *protocol.ReLoginReq) (*protocol.ReLoginRsp, int32) {
	if req.Session == "" {
		return nil, code.SessionInvalid
	}

	ctx := context.Background()
	key := fmt.Sprintf("session:%s", req.Session)
	uid, err := p.redisClient.Get(ctx, key).Int64()
	if err != nil || uid == 0 {
		return nil, code.SessionInvalid
	}

	newToken := p.generateToken(uid, "")
	p.redisClient.Set(ctx, fmt.Sprintf("token:%s", newToken), uid, 24*time.Hour)

	return &protocol.ReLoginRsp{Session: newToken}, code.OK
}

func (p *ActorAccount) logout(req *protocol.LogoutReq) (*protocol.LogoutRsp, int32) {
	if req.UId > 0 {
		ctx := context.Background()
		p.redisClient.Del(ctx, fmt.Sprintf("uid_online:%d", req.UId))
	}
	return &protocol.LogoutRsp{UId: req.UId}, code.OK
}

func (p *ActorAccount) serverList(_ *protocol.ServerListReq) (*protocol.ServerListRsp, int32) {
	servers := []protocol.Server{
		{Id: 1, Slg: "ws://127.0.0.1:8004", Chat: "ws://127.0.0.1:8005"},
	}
	return &protocol.ServerListRsp{Lists: servers}, code.OK
}

type changePwdReq struct {
	Username string `json:"username"`
	OldPwd   string `json:"oldPwd"`
	NewPwd   string `json:"newPwd"`
}

func (p *ActorAccount) changePwd(req *changePwdReq) int32 {
	acc, err := p.accountEntry.FindByUsername(req.Username)
	if err != nil || acc == nil {
		return code.UserNotExist
	}
	if acc.Password != req.OldPwd {
		return code.PwdIncorrect
	}
	acc.Password = req.NewPwd
	if err := p.accountEntry.Update(acc); err != nil {
		return code.DBError
	}
	return code.OK
}

type forgetPwdReq struct {
	Username string `json:"username"`
}

type forgetPwdRsp struct {
	Username string `json:"username"`
}

func (p *ActorAccount) forgetPwd(req *forgetPwdReq) (*forgetPwdRsp, int32) {
	acc, err := p.accountEntry.FindByUsername(req.Username)
	if err != nil || acc == nil {
		return nil, code.UserNotExist
	}
	return &forgetPwdRsp{Username: acc.Username}, code.OK
}

type resetPwdReq struct {
	Username string `json:"username"`
	NewPwd   string `json:"newPwd"`
}

func (p *ActorAccount) resetPwd(req *resetPwdReq) int32 {
	acc, err := p.accountEntry.FindByUsername(req.Username)
	if err != nil || acc == nil {
		return code.UserNotExist
	}
	acc.Password = req.NewPwd
	if err := p.accountEntry.Update(acc); err != nil {
		return code.DBError
	}
	return code.OK
}

func (p *ActorAccount) generateToken(uid int64, username string) string {
	data := fmt.Sprintf("%d:%s:%d", uid, username, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}
