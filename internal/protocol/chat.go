package protocol

type ChatMsg struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
	Type     int8   `json:"type"`
	Msg      string `json:"msg"`
	Time     int64  `json:"time"`
}

type ChatLoginReq struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
	Token    string `json:"token"`
}

type ChatLoginRsp struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
}

type ChatLogoutReq struct {
	RId int `json:"RId"`
}

type ChatLogoutRsp struct {
	RId int `json:"RId"`
}

type ChatReq struct {
	Type int8   `json:"type"`
	Msg  string `json:"msg"`
}

type ChatHistoryReq struct {
	Type int8 `json:"type"`
}

type ChatHistoryRsp struct {
	Type int8      `json:"type"`
	Msgs []ChatMsg `json:"msgs"`
}

type ChatJoinReq struct {
	Type int8 `json:"type"`
	Id   int  `json:"id"`
}

type ChatJoinRsp struct {
	Type int8 `json:"type"`
	Id   int  `json:"id"`
}

type ChatExitReq struct {
	Type int8 `json:"type"`
	Id   int  `json:"id"`
}

type ChatExitRsp struct {
	Type int8 `json:"type"`
	Id   int  `json:"id"`
}
