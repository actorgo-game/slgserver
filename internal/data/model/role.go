package model

import (
	"time"

	"github.com/llr104/slgserver/internal/protocol"
)

type Role struct {
	RId        int       `bson:"rid" json:"rid"`
	ServerId   int       `bson:"server_id" json:"serverId"`
	UId        int       `bson:"uid" json:"uid"`
	NickName   string    `bson:"nick_name" json:"nickName"`
	Balance    int       `bson:"balance" json:"balance"`
	HeadId     int16     `bson:"head_id" json:"headId"`
	Sex        int8      `bson:"sex" json:"sex"`
	Profile    string    `bson:"profile" json:"profile"`
	LoginTime  time.Time `bson:"login_time" json:"loginTime"`
	LogoutTime time.Time `bson:"logout_time" json:"logoutTime"`
	CreatedAt  time.Time `bson:"created_at" json:"createdAt"`
}

func (Role) CollectionName() string {
	return "roles"
}

func (this *Role) ToProto() interface{} {
	p := protocol.Role{}
	p.UId = this.UId
	p.RId = this.RId
	p.Sex = this.Sex
	p.NickName = this.NickName
	p.HeadId = this.HeadId
	p.Balance = this.Balance
	p.Profile = this.Profile
	return p
}
