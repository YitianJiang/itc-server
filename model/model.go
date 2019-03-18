package model

/*
lark返回的user结构提
*/
type LarkUserStruct struct {
	Ok      bool   `json:"ok"`
	User_id string `json:"user_id"`
}
type IMStruct struct {
	Channel IMChannelStruct `json:"channel"`
	Ok      bool            `json:"ok"`
}

type IMChannelStruct struct {
	Id string `json:"id"`
}
