package form

type GetChatInfoForm struct {
	OpenChatId    *string         		`json:"open_chat_id" validate:"required"`
}

type GetChatListForm struct {
	Page        *int16       			`json:"page" validate:"required"`
	PageSize    *int8		 			`json:"page_size" validate:"omitempty"`
}

type CreateChatForm struct {
	Name          *string       		`json:"name" validate:"omitempty"`
	Description   *string      			`json:"description" validate:"omitempty"`
	OpenIds      []*string   			`json:"open_ids" validate:"required"`
}

type User2ChatForm struct {
	OpenChatId    *string               `json:"open_chat_id" validate:"required"`
	OpenIds       []*string             `json:"open_ids" validate:"required"`
}

type UpdateChatInfoForm struct {
	OpenChatId   *string                `json:"open_chat_id" validate:"required"`
	OwnerId      *string                `json:"owner_id" validate:"omitempty"`
	Name         *string                `json:"name" validate:"omitempty"`
	I18nNames    map[string]string		`json:"i18n_names" validate:"omitempty"`
}

type GetP2pChatIdForm struct {
	UserId      *string                 `json:"user_id" validate:"omitempty"`
	OpenId      *string                 `json:"open_id" validate:"omitempty"`
	Chatter     *string                 `json:"chatter" validate:"omitempty"`
}
//解散群
type DisbandChatForm struct {
	ChatId      *string					`json:"chat_id" validate:"omitempty"`
}