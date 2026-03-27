package model

const (
	ArmyCmdIdle        = 0
	ArmyCmdAttack      = 1
	ArmyCmdDefend      = 2
	ArmyCmdReclamation = 3
	ArmyCmdBack        = 4
	ArmyCmdConscript   = 5
	ArmyCmdTransfer    = 6
)

type Army struct {
	Id             int   `bson:"id" json:"id"`
	ServerId       int   `bson:"server_id" json:"serverId"`
	RId            int   `bson:"rid" json:"rid"`
	CityId         int   `bson:"cityId" json:"cityId"`
	Order          int8  `bson:"order" json:"order"`
	Generals       [3]int `bson:"generals" json:"generals"`
	Soldiers       [3]int `bson:"soldiers" json:"soldiers"`
	ConscriptTimes [3]int64 `bson:"conscript_times" json:"conscriptTimes"`
	ConscriptCnts  [3]int  `bson:"conscript_cnts" json:"conscriptCnts"`
	Cmd            int8  `bson:"cmd" json:"cmd"`
	FromX          int   `bson:"from_x" json:"fromX"`
	FromY          int   `bson:"from_y" json:"fromY"`
	ToX            int   `bson:"to_x" json:"toX"`
	ToY            int   `bson:"to_y" json:"toY"`
	Start          int64 `bson:"start" json:"start"`
	End            int64 `bson:"end" json:"end"`
}

func (Army) CollectionName() string {
	return "armies"
}
