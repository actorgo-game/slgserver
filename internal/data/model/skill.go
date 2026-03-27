package model

type Skill struct {
	Id             int   `bson:"id" json:"id"`
	ServerId       int   `bson:"server_id" json:"serverId"`
	RId            int   `bson:"rid" json:"rid"`
	CfgId          int   `bson:"cfgId" json:"cfgId"`
	BelongGenerals []int `bson:"belong_generals" json:"belongGenerals"`
}

func (Skill) CollectionName() string {
	return "skills"
}
