package model

type RoleRes struct {
	Id       int `bson:"id" json:"id"`
	ServerId int `bson:"server_id" json:"serverId"`
	RId      int `bson:"rid" json:"rid"`
	Wood     int `bson:"wood" json:"wood"`
	Iron     int `bson:"iron" json:"iron"`
	Stone    int `bson:"stone" json:"stone"`
	Grain    int `bson:"grain" json:"grain"`
	Gold     int `bson:"gold" json:"gold"`
	Decree   int `bson:"decree" json:"decree"`
}

func (RoleRes) CollectionName() string {
	return "role_resources"
}
