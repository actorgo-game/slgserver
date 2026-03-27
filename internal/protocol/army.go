package protocol

type Army struct {
	Id       int      `json:"id"`
	CityId   int      `json:"cityId"`
	UnionId  int      `json:"union_id"`
	Order    int8     `json:"order"`
	Generals [3]int   `json:"generals"`
	Soldiers [3]int   `json:"soldiers"`
	ConTimes [3]int64 `json:"con_times"`
	ConCnts  [3]int   `json:"con_cnts"`
	Cmd      int8     `json:"cmd"`
	State    int8     `json:"state"`
	FromX    int      `json:"from_x"`
	FromY    int      `json:"from_y"`
	ToX      int      `json:"to_x"`
	ToY      int      `json:"to_y"`
	Start    int64    `json:"start"`
	End      int64    `json:"end"`
}

type ArmyListReq struct {
	CityId int `json:"cityId"`
}

type ArmyListRsp struct {
	CityId int    `json:"cityId"`
	Armys  []Army `json:"armys"`
}

type ArmyOneReq struct {
	CityId int  `json:"cityId"`
	Order  int8 `json:"order"`
}

type ArmyOneRsp struct {
	Army Army `json:"army"`
}

type DisposeReq struct {
	CityId    int  `json:"cityId"`
	GeneralId int  `json:"generalId"`
	Order     int8 `json:"order"`
	Position  int  `json:"position"`
}

type DisposeRsp struct {
	Army Army `json:"army"`
}

type ConscriptReq struct {
	ArmyId int   `json:"armyId"`
	Cnts   []int `json:"cnts"`
}

type ConscriptRsp struct {
	Army    Army    `json:"army"`
	RoleRes RoleRes `json:"role_res"`
}

type AssignArmyReq struct {
	ArmyId int  `json:"armyId"`
	Cmd    int8 `json:"cmd"`
	X      int  `json:"x"`
	Y      int  `json:"y"`
}

type AssignArmyRsp struct {
	Army Army `json:"army"`
}
