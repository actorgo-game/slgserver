package httpserver

import (
	"time"

	cgin "github.com/actorgo-game/actorgo/components/gin"
	cmongo "github.com/actorgo-game/actorgo/components/mongo"
	cerror "github.com/actorgo-game/actorgo/error"
	clog "github.com/actorgo-game/actorgo/logger"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/entry"
	"github.com/llr104/slgserver/internal/data/model"
)

type Controller struct {
	cgin.BaseController
	accountEntry *entry.AccountEntry
}

func (p *Controller) Init() {
	db := cmongo.Instance().GetDb("slg_db")
	if db == nil {
		clog.Error("[Controller] slg_db not found")
		return
	}
	p.accountEntry = entry.NewAccountEntry(db.Collection("accounts"))

	group := p.Group("/")
	group.GET("/account/register", p.handleRegister)
	group.GET("/account/changepwd", p.handleChangePwd)
	group.GET("/account/forgetpwd", p.handleForgetPwd)
	group.GET("/account/resetpwd", p.handleResetPwd)
}

func (p *Controller) handleRegister(c *cgin.Context) {
	clog.Info("register")
	data := make(map[string]interface{})

	statusCode := code.OK
	err, statusCode := p.CreateUser(c)

	// 错误处理 + 错误日志（关键）
	if err != nil {
		clog.Error("user register failed, err: %v, statusCode: %v", err, statusCode)
		data["errmsg"] = err.Error()
	}

	// 统一赋值状态码
	data["code"] = statusCode
	// 统一返回结果
	code.RenderResult(c, statusCode, data)
}

func (p *Controller) handleChangePwd(c *cgin.Context) {
	clog.Info("changePwd")
	data := make(map[string]interface{})

	statusCode := code.OK
	err, statusCode := p.ChangePassword(c)

	if err != nil {
		clog.Error("user register failed, err: %v, statusCode: %v", err, statusCode)
		data["errmsg"] = err.Error()
	}

	// 统一赋值状态码
	data["code"] = statusCode
	// 统一返回结果
	code.RenderResult(c, statusCode, data)
}

func (p *Controller) handleForgetPwd(c *cgin.Context) {
	clog.Info("forgetPwd")
	code.RenderResult(c, code.OK)
}

func (p *Controller) handleResetPwd(c *cgin.Context) {
	clog.Info("forgetPwd")
	code.RenderResult(c, code.OK)
}

func (p *Controller) CreateUser(ctx *cgin.Context) (error, int32) {
	account := ctx.GetString("username", "", true)
	pwd := ctx.GetString("password", "", true)

	if len(account) == 0 || len(pwd) == 0 {
		return cerror.Error("用户名或密码是空"), code.InvalidParam
	}

	if p.UserExists(account) {
		return cerror.Error("账号已经存在"), code.UserExist
	}

	acc := &model.Account{
		AccountId:  time.Now().UnixMilli(),
		Username:   account,
		Password:   pwd,
		CreateTime: time.Now(),
	}

	if err := p.accountEntry.Insert(acc); err != nil {
		return cerror.Error("数据库出错"), code.DBError
	}
	return nil, code.OK
}

func (p *Controller) ChangePassword(ctx *cgin.Context) (error, int32) {
	account := ctx.GetString("username", "", true)
	pwd := ctx.GetString("password", "", true)
	newpwd := ctx.GetString("newpassword", "", true)

	if len(account) == 0 || len(pwd) == 0 || len(newpwd) == 0 {
		return cerror.Error("用户名或密码是空"), code.InvalidParam
	}

	acc, err := p.accountEntry.FindByUsername(account)
	if err != nil || acc == nil {
		return cerror.Error("用户不存在"), code.UserNotExist
	}

	if acc.Password != pwd {
		return cerror.Error("原密码错误"), code.PwdIncorrect
	}

	acc.Password = newpwd
	if err := p.accountEntry.Update(acc); err != nil {
		return cerror.Error("数据库出错"), code.DBError
	}
	return nil, code.OK
}

func (p *Controller) UserExists(username string) bool {
	acc, err := p.accountEntry.FindByUsername(username)
	return err == nil && acc != nil
}
