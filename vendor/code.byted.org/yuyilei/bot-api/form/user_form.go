package form

type GetUserInfoForm struct {
	OpenId    *string    `json:"open_id" validate:"required"`
}

type GetAdminInfoListForm struct {
	OpenId    *string    `json:"open_id" validate:"omitempty"`
}

//拉踢机器人结构
type BotForm struct {
	ChatId    *string	 `json:"chat_id" validate:"omitempty"`
}

