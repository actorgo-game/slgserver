package model

import "time"

type Role struct {
	RId        int       `bson:"rid" json:"rid"`
	ServerId   int       `bson:"server_id" json:"serverId"`
	UId        int       `bson:"uid" json:"uid"`
	NickName   string    `bson:"nick_name" json:"nickName"`
	Balance    int       `bson:"balance" json:"balance"`
	HeadId     int       `bson:"head_id" json:"headId"`
	Sex        int8      `bson:"sex" json:"sex"`
	Profile    string    `bson:"profile" json:"profile"`
	LoginTime  time.Time `bson:"login_time" json:"loginTime"`
	LogoutTime time.Time `bson:"logout_time" json:"logoutTime"`
	CreatedAt  time.Time `bson:"created_at" json:"createdAt"`
}

func (Role) CollectionName() string {
	return "roles"
}
