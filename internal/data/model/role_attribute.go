package model

type PosTag struct {
	X    int    `bson:"x" json:"x"`
	Y    int    `bson:"y" json:"y"`
	Name string `bson:"name" json:"name"`
}

type RoleAttribute struct {
	Id              int      `bson:"id" json:"id"`
	ServerId        int      `bson:"server_id" json:"serverId"`
	RId             int      `bson:"rid" json:"rid"`
	UnionId         int      `bson:"union_id" json:"unionId"`
	ParentId        int      `bson:"parent_id" json:"parentId"`
	CollectTimes    int      `bson:"collect_times" json:"collectTimes"`
	LastCollectTime int64    `bson:"last_collect_time" json:"lastCollectTime"`
	PosTags         []PosTag `bson:"pos_tags" json:"posTags"`
}

func (RoleAttribute) CollectionName() string {
	return "role_attributes"
}
