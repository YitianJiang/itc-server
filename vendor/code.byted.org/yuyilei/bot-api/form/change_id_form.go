package form


type OpenId2UserIdForm struct {
	OpenId    *string    `json:"open_id" validate:"required"`
}

type UserId2OpenIdForm struct {
	UserId    *string    `json:"user_id" validate:"required"`
}

type MessageId2OpenMessageIdForm struct {
	MessageId   *string   `json:"message_id" validate:"required"`
}

type OpenMessageId2MessageIdForm struct {
	OpenMessageId   *string    `json:"open_message_id" validate:"required"`
}

type ChatId2OpenChatIdForm struct {
	ChatId   *string    `json:"chat_id" validate:"required"`
}

type OpenChatId2ChatIdForm struct {
	OpenChatId   *string    `json:"open_chat_id" validate:"required"`
}

type UserId2EmployeeIdForm struct {
	UserId    *string     `json:"user_id" validate:"required"`
}

type EmployeeId2UserIdForm struct {
	EmployeeId   *string    `json:"employee_id" validate:"required"`
}

type DepartmentId2OpenDepartmentIdForm struct {
	DepartmentId    *string     `json:"department_id" validate:"required"`
}

type OpenDepartmentId2DepartmentIdForm struct {
	OpenDepartmentId   *string    `json:"open_department_id" validate:"required"`
}

type Email2OpenIdEmployeeIdUrlForm struct {
	Email      string     `json:"email" validate:"required"`
}