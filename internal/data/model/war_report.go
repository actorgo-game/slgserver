package model

import "time"

type WarReport struct {
	Id               int       `bson:"id" json:"id"`
	ServerId         int       `bson:"server_id" json:"serverId"`
	AttackRid        int       `bson:"a_rid" json:"a_rid"`
	DefenseRid       int       `bson:"d_rid" json:"d_rid"`
	BegAttackArmy    string    `bson:"beg_a_army" json:"begAttackArmy"`
	BegDefenseArmy   string    `bson:"beg_d_army" json:"begDefenseArmy"`
	EndAttackArmy    string    `bson:"end_a_army" json:"endAttackArmy"`
	EndDefenseArmy   string    `bson:"end_d_army" json:"endDefenseArmy"`
	BegAttackGeneral string    `bson:"beg_a_general" json:"begAttackGeneral"`
	BegDefenseGeneral string   `bson:"beg_d_general" json:"begDefenseGeneral"`
	EndAttackGeneral string    `bson:"end_a_general" json:"endAttackGeneral"`
	EndDefenseGeneral string   `bson:"end_d_general" json:"endDefenseGeneral"`
	Result           int8      `bson:"result" json:"result"`
	Rounds           int       `bson:"rounds" json:"rounds"`
	AttackIsRead     bool      `bson:"a_is_read" json:"aIsRead"`
	DefenseIsRead    bool      `bson:"d_is_read" json:"dIsRead"`
	DestroyDurable   int       `bson:"destroy_durable" json:"destroyDurable"`
	Occupy           int       `bson:"occupy" json:"occupy"`
	X                int       `bson:"x" json:"x"`
	Y                int       `bson:"y" json:"y"`
	CreatedAt        time.Time `bson:"created_at" json:"createdAt"`
}

func (WarReport) CollectionName() string {
	return "war_reports"
}
