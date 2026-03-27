package model

import "time"

const (
	GeneralNormal   = 0
	GeneralComposed = 1
	GeneralConverted = 2
)

type GeneralSkill struct {
	Id    int `bson:"id" json:"id"`
	Lv    int `bson:"lv" json:"lv"`
	CfgId int `bson:"cfgId" json:"cfgId"`
}

type General struct {
	Id            int            `bson:"id" json:"id"`
	ServerId      int            `bson:"server_id" json:"serverId"`
	RId           int            `bson:"rid" json:"rid"`
	CfgId         int            `bson:"cfgId" json:"cfgId"`
	PhysicalPower int            `bson:"physical_power" json:"physicalPower"`
	Level         int            `bson:"level" json:"level"`
	Exp           int            `bson:"exp" json:"exp"`
	Order         int8           `bson:"order" json:"order"`
	CityId        int            `bson:"cityId" json:"cityId"`
	CurArms       int            `bson:"arms" json:"curArms"`
	HasPrPoint    int            `bson:"has_pr_point" json:"hasPrPoint"`
	UsePrPoint    int            `bson:"use_pr_point" json:"usePrPoint"`
	AttackDis     int            `bson:"attack_distance" json:"attackDis"`
	ForceAdded    int            `bson:"force_added" json:"forceAdded"`
	StrategyAdded int            `bson:"strategy_added" json:"strategyAdded"`
	DefenseAdded  int            `bson:"defense_added" json:"defenseAdded"`
	SpeedAdded    int            `bson:"speed_added" json:"speedAdded"`
	DestroyAdded  int            `bson:"destroy_added" json:"destroyAdded"`
	StarLv        int8           `bson:"star_lv" json:"starLv"`
	Star          int8           `bson:"star" json:"star"`
	ParentId      int            `bson:"parentId" json:"parentId"`
	Skills        []GeneralSkill `bson:"skills" json:"skills"`
	State         int8           `bson:"state" json:"state"`
	CreatedAt     time.Time      `bson:"created_at" json:"createdAt"`
}

func (General) CollectionName() string {
	return "generals"
}
