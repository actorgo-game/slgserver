package model

import "time"

const (
	UnionDismiss = 0
	UnionRunning = 1
)

type Coalition struct {
	Id            int       `bson:"id" json:"id"`
	ServerId      int       `bson:"server_id" json:"serverId"`
	Name          string    `bson:"name" json:"name"`
	Members       []int     `bson:"members" json:"members"`
	CreateId      int       `bson:"create_id" json:"createId"`
	Chairman      int       `bson:"chairman" json:"chairman"`
	ViceChairman  int       `bson:"vice_chairman" json:"viceChairman"`
	Notice        string    `bson:"notice" json:"notice"`
	State         int8      `bson:"state" json:"state"`
	CreatedAt     time.Time `bson:"created_at" json:"createdAt"`
}

func (Coalition) CollectionName() string {
	return "coalitions"
}

type CoalitionApply struct {
	Id        int       `bson:"id" json:"id"`
	ServerId  int       `bson:"server_id" json:"serverId"`
	UnionId   int       `bson:"union_id" json:"unionId"`
	RId       int       `bson:"rid" json:"rid"`
	State     int8      `bson:"state" json:"state"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

func (CoalitionApply) CollectionName() string {
	return "coalition_applies"
}

type CoalitionLog struct {
	Id        int       `bson:"id" json:"id"`
	ServerId  int       `bson:"server_id" json:"serverId"`
	UnionId   int       `bson:"union_id" json:"unionId"`
	OPRId     int       `bson:"op_rid" json:"opRId"`
	TargetId  int       `bson:"target_id" json:"targetId"`
	State     int8      `bson:"state" json:"state"`
	Des       string    `bson:"des" json:"des"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

func (CoalitionLog) CollectionName() string {
	return "coalition_logs"
}
