package protocol

type CollectionReq struct{}

type CollectionRsp struct {
	Gold     int   `json:"gold"`
	Limit    int8  `json:"limit"`
	CurTimes int8  `json:"cur_times"`
	NextTime int64 `json:"next_time"`
}

type OpenCollectionReq struct{}

type OpenCollectionRsp struct {
	Limit    int8  `json:"limit"`
	CurTimes int8  `json:"cur_times"`
	NextTime int64 `json:"next_time"`
}

type TransformReq struct {
	From []int `json:"from"`
	To   []int `json:"to"`
}

type TransformRsp struct{}
