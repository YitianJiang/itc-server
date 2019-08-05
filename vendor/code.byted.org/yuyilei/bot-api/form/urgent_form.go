package form

type UrgentMessageForm struct {
	OpenMessageID string   `json:"open_message_id" validate:"required"`
	UrgentType    string   `json:"urgent_type" validate:"required"`
	OpenIDs       []string `json:"open_ids" validate:"required,min=1,max=100"`
}