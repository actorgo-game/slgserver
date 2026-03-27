package model

import "time"

type Account struct {
	AccountId  int64     `bson:"account_id" json:"accountId"`
	Username   string    `bson:"username" json:"username"`
	Password   string    `bson:"password" json:"-"`
	CreateIP   string    `bson:"create_ip" json:"createIP"`
	CreateTime time.Time `bson:"create_time" json:"createTime"`
}

func (Account) CollectionName() string {
	return "accounts"
}
