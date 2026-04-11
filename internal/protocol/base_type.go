package protocol

type StringKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LoginRequest struct {
	ServerId int32            `json:"serverId"` // 当前登陆的服务器id
	Token    string           `json:"token"`    // 登陆token(web login api生成的base64字符串)
	Params   map[int32]string `json:"params" `  // 登陆时上传的参数 key: LoginParams
}

type LoginResponse struct {
	Uid    int64            `json:"uid"`     // 游戏内的用户唯一id
	Pid    int32            `json:"pid"`     // 平台id
	OpenId string           `json:"openId"`  // 平台openId(平台的帐号唯一id)
	Params map[int32]string `json:"params" ` // 登陆后的扩展参数，按需增加
}
