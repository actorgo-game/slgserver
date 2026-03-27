package protocol

const (
	UnionChairman     = 0
	UnionViceChairman = 1
	UnionCommon       = 2
)

const (
	UnionUntreated = 0
	UnionRefuse    = 1
	UnionAdopt     = 2
)

type Member struct {
	RId   int    `json:"rid"`
	Name  string `json:"name"`
	Title int8   `json:"title"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

type Major struct {
	RId   int    `json:"rid"`
	Name  string `json:"name"`
	Title int8   `json:"title"`
}

type Union struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Cnt    int     `json:"cnt"`
	Notice string  `json:"notice"`
	Major  []Major `json:"major"`
}

type ApplyItem struct {
	Id       int    `json:"id"`
	RId      int    `json:"rid"`
	NickName string `json:"nick_name"`
}

type CreateReq struct {
	Name string `json:"name"`
}

type CreateRsp struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ListReq struct{}

type ListRsp struct {
	List []Union `json:"list"`
}

type JoinReq struct {
	Id int `json:"id"`
}

type JoinRsp struct{}

type MemberReq struct {
	Id int `json:"id"`
}

type MemberRsp struct {
	Id      int      `json:"id"`
	Members []Member `json:"Members"`
}

type ApplyReq struct {
	Id int `json:"id"`
}

type ApplyRsp struct {
	Id     int         `json:"id"`
	Applys []ApplyItem `json:"applys"`
}

type VerifyReq struct {
	Id     int  `json:"id"`
	Decide int8 `json:"decide"`
}

type VerifyRsp struct {
	Id     int  `json:"id"`
	Decide int8 `json:"decide"`
}

type ExitReq struct{}

type ExitRsp struct{}

type DismissReq struct{}

type DismissRsp struct{}

type NoticeReq struct {
	Id int `json:"id"`
}

type NoticeRsp struct {
	Text string `json:"text"`
}

type ModNoticeReq struct {
	Text string `json:"text"`
}

type ModNoticeRsp struct {
	Id   int    `json:"id"`
	Text string `json:"text"`
}

type KickReq struct {
	RId int `json:"rid"`
}

type KickRsp struct {
	RId int `json:"rid"`
}

type AppointReq struct {
	RId   int `json:"rid"`
	Title int `json:"title"`
}

type AppointRsp struct {
	RId   int `json:"rid"`
	Title int `json:"title"`
}

type AbdicateReq struct {
	RId int `json:"rid"`
}

type AbdicateRsp struct{}

type InfoReq struct {
	Id int `json:"id"`
}

type InfoRsp struct {
	Id   int   `json:"id"`
	Info Union `json:"info"`
}

type UnionLog struct {
	OPRId    int    `json:"op_rid"`
	TargetId int    `json:"target_id"`
	State    int8   `json:"state"`
	Des      string `json:"des"`
	Ctime    int64  `json:"ctime"`
}

type LogReq struct{}

type LogRsp struct {
	Logs []UnionLog `json:"logs"`
}
