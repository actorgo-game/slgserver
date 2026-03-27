package protocol

type GSkill struct {
	Id    int `json:"id"`
	Lv    int `json:"lv"`
	CfgId int `json:"cfgId"`
}

type General struct {
	Id            int      `json:"id"`
	CfgId         int      `json:"cfgId"`
	PhysicalPower int      `json:"physical_power"`
	Order         int8     `json:"order"`
	Level         int8     `json:"level"`
	Exp           int      `json:"exp"`
	CityId        int      `json:"cityId"`
	CurArms       int      `json:"curArms"`
	HasPrPoint    int      `json:"hasPrPoint"`
	UsePrPoint    int      `json:"usePrPoint"`
	AttackDis     int      `json:"attack_distance"`
	ForceAdded    int      `json:"force_added"`
	StrategyAdded int      `json:"strategy_added"`
	DefenseAdded  int      `json:"defense_added"`
	SpeedAdded    int      `json:"speed_added"`
	DestroyAdded  int      `json:"destroy_added"`
	StarLv        int8     `json:"star_lv"`
	Star          int8     `json:"star"`
	ParentId      int      `json:"parentId"`
	Skills        []*GSkill `json:"skills"`
	State         int8     `json:"state"`
}

type MyGeneralReq struct{}

type MyGeneralRsp struct {
	Generals []General `json:"generals"`
}

type DrawGeneralReq struct {
	DrawTimes int `json:"drawTimes"`
}

type DrawGeneralRsp struct {
	Generals []General `json:"generals"`
}

type ComposeGeneralReq struct {
	CompId int   `json:"compId"`
	GIds   []int `json:"gIds"`
}

type ComposeGeneralRsp struct {
	Generals []General `json:"generals"`
}

type AddPrGeneralReq struct {
	CompId      int `json:"compId"`
	ForceAdd    int `json:"forceAdd"`
	StrategyAdd int `json:"strategyAdd"`
	DefenseAdd  int `json:"defenseAdd"`
	SpeedAdd    int `json:"speedAdd"`
	DestroyAdd  int `json:"destroyAdd"`
}

type AddPrGeneralRsp struct {
	Generals General `json:"general"`
}

type ConvertReq struct {
	GIds []int `json:"gIds"`
}

type ConvertRsp struct {
	GIds    []int `json:"gIds"`
	Gold    int   `json:"gold"`
	AddGold int   `json:"add_gold"`
}

type UpDownSkillReq struct {
	GId   int `json:"gId"`
	CfgId int `json:"cfgId"`
	Pos   int `json:"pos"`
}

type UpDownSkillRsp struct {
	GId   int `json:"gId"`
	CfgId int `json:"cfgId"`
	Pos   int `json:"pos"`
}

type LvSkillReq struct {
	GId int `json:"gId"`
	Pos int `json:"pos"`
}

type LvSkillRsp struct {
	GId int `json:"gId"`
	Pos int `json:"pos"`
}
