package protocol

type Conf struct {
	Type     int8   `json:"type"`
	Level    int8   `json:"level"`
	Name     string `json:"name"`
	Wood     int    `json:"Wood"`
	Iron     int    `json:"iron"`
	Stone    int    `json:"stone"`
	Grain    int    `json:"grain"`
	Durable  int    `json:"durable"`
	Defender int    `json:"defender"`
}

type ConfigReq struct{}

type ConfigRsp struct {
	Confs []Conf `json:"confs"`
}

type MapRoleBuild struct {
	RId        int    `json:"rid"`
	RNick      string `json:"RNick"`
	Name       string `json:"name"`
	UnionId    int    `json:"union_id"`
	UnionName  string `json:"union_name"`
	ParentId   int    `json:"parent_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Type       int8   `json:"type"`
	Level      int8   `json:"level"`
	OPLevel    int8   `json:"op_level"`
	CurDurable int    `json:"cur_durable"`
	MaxDurable int    `json:"max_durable"`
	Defender   int    `json:"defender"`
	OccupyTime int64  `json:"occupy_time"`
	EndTime    int64  `json:"end_time"`
	GiveUpTime int64  `json:"giveUp_time"`
}

type ScanReq struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ScanRsp struct {
	MRBuilds []MapRoleBuild `json:"mr_builds"`
	MCBuilds []MapRoleCity  `json:"mc_builds"`
	Armys    []Army         `json:"armys"`
}

type ScanBlockReq struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Length int `json:"length"`
}

type GiveUpReq struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type GiveUpRsp struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type BuildReq struct {
	X    int  `json:"x"`
	Y    int  `json:"y"`
	Type int8 `json:"type"`
}

type BuildRsp struct {
	X    int  `json:"x"`
	Y    int  `json:"y"`
	Type int8 `json:"type"`
}

type UpBuildReq struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type UpBuildRsp struct {
	X     int          `json:"x"`
	Y     int          `json:"y"`
	Build MapRoleBuild `json:"build"`
}

type DelBuildReq struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type DelBuildRsp struct {
	X     int          `json:"x"`
	Y     int          `json:"y"`
	Build MapRoleBuild `json:"build"`
}
