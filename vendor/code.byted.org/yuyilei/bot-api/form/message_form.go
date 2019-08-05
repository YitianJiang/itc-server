package form

type MessageContent struct {
	Text            *string                  `json:"text,omitempty" validate:"omitempty"`
	Title           *string                  `json:"title,omitempty" validate:"omitempty"`
	ImageKey        *string                  `json:"image_key,omitempty" validate:"omitempty"`
	Card            CardForm                 `json:"card,omitempty" validate:"omitempty"`
	Post            map[string]*PostForm     `json:"post,omitempty" validate:"omitempty"`
	ShareOpenChatID *string                  `json:"share_open_chat_id,omitempty" validate:"omitempty"`
}

type SendMessageForm struct {
	OpenID     *string        				  `json:"open_id,omitempty" validate:"omitempty"`
	UserID     *string        				  `json:"user_id,omitempty" validate:"omitempty"`
	ChatID     *string        				  `json:"chat_id,omitempty" validate:"omitempty"`
	EmployeeID *string        				  `json:"employee_id,omitempty" validate:"omitempty"`
	OpenChatID *string        				  `json:"open_chat_id,omitempty" validate:"omitempty"`
	RootID     *string        				  `json:"root_id,omitempty" validate:"omitempty"`
	Email      *string        				  `json:"email,omitempty" validate:"omitempty"`
	MsgType    *string         				  `json:"msg_type,omitempty" validate:"required"`
	Content    MessageContent  				  `json:"content,omitempty" validate:"required"`
	DepartmentIds  []string     			  `json:"department_ids,omitempty" validate:"omitempty"`
	OpenIds   []string                        `json:"open_ids,omitempty" validate:"omitempty"`
}

type SendMessageNativeForm struct {
	BotID     string         `json:"bot_id" validate:"required"`
	UserID    string         `json:"user_id" validate:"required"`
	CreatorId string         `json:"creator_id" validate:"required"`
	MsgType   string         `json:"msg_type" validate:"required"`
	Content   MessageContent `json:"content" validate:"required"`
}