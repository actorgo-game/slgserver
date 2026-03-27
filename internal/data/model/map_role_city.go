package model

import "time"

type MapRoleCity struct {
	CityId     int       `bson:"cityId" json:"cityId"`
	ServerId   int       `bson:"server_id" json:"serverId"`
	RId        int       `bson:"rid" json:"rid"`
	Name       string    `bson:"name" json:"name"`
	X          int       `bson:"x" json:"x"`
	Y          int       `bson:"y" json:"y"`
	IsMain     bool      `bson:"is_main" json:"isMain"`
	CurDurable int       `bson:"cur_durable" json:"curDurable"`
	MaxDurable int       `bson:"max_durable" json:"maxDurable"`
	CreatedAt  time.Time `bson:"created_at" json:"createdAt"`
	OccupyTime time.Time `bson:"occupy_time" json:"occupyTime"`
}

func (MapRoleCity) CollectionName() string {
	return "cities"
}
