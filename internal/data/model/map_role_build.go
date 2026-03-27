package model

import "time"

const (
	MapBuildSysFortress = 50
	MapBuildSysCity     = 51
	MapBuildFortress    = 56
)

type MapRoleBuild struct {
	Id         int       `bson:"id" json:"id"`
	ServerId   int       `bson:"server_id" json:"serverId"`
	RId        int       `bson:"rid" json:"rid"`
	Type       int8      `bson:"type" json:"type"`
	Level      int8      `bson:"level" json:"level"`
	OPLevel    int8      `bson:"op_level" json:"opLevel"`
	X          int       `bson:"x" json:"x"`
	Y          int       `bson:"y" json:"y"`
	Name       string    `bson:"name" json:"name"`
	CurDurable int       `bson:"cur_durable" json:"curDurable"`
	MaxDurable int       `bson:"max_durable" json:"maxDurable"`
	OccupyTime time.Time `bson:"occupy_time" json:"occupyTime"`
	EndTime    time.Time `bson:"end_time" json:"endTime"`
	GiveUpTime int64     `bson:"give_up_time" json:"giveUpTime"`
}

func (MapRoleBuild) CollectionName() string {
	return "builds"
}
