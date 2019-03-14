package lark

// BaseResp ...
type BaseResp struct {
	Code  int    `json:"code,omitempty"`
	Ok    bool   `json:"ok,omitempty"`
	Error string `json:"error,omitempty"`
}

// PostMessageResp response of PostMessage
type PostMessageResp struct {
	BaseResp
	Timestamp     float64        `json:"ts,omitempty"`
	ChatID        string         `json:"chat_id,omitempty"`
	OpenMessageID string         `json:"open_message_id,omitempty"`
	MsgType       string         `json:"msg_type,omitempty"`
	Message       MessageContent `json:"message,omitempty"`
}

// PostPrivateMessageResp is the same as PostMessageResp by now
type PostPrivateMessageResp PostMessageResp

// UserInfoByEmailResp user info by email struct
type UserInfoByEmailResp struct {
	BaseResp
	UserID string `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
}

// UserInfoResp user info struct
type UserInfoResp struct {
	BaseResp
	UserID     string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	EnName     string `json:"en_name,omitempty"`
	ProfilePic string `json:"profile_pic,omitempty"`
}

// BotInfo bot info
type BotInfo struct {
	AvatarKey string `json:"avatar_key,omitempty"`
	Name      string `json:"name,omitempty"`
	Token     string `json:"token,omitempty"`
}

// BotInfoResp response of bot info
type BotInfoResp struct {
	BaseResp
	Bot BotInfo `json:"bot,omitempty"`
}

// BotListResp response of bot list
type BotListResp struct {
	BaseResp
	Bots []BotInfo `json:"bots,omitempty"`
}

// AddEventResp response of add event
type AddEventResp BaseResp

// AddEventHookResp response of add event
type AddEventHookResp BaseResp

// UploadImageResp response of upload image
type UploadImageResp struct {
	BaseResp
	ImageKey string `json:"image_key,omitempty"`
	URL      string `json:"url,omitempty"`
}

// ChannelInfo ...
type ChannelInfo struct {
	ID       string   `json:"ID,omitempty"`
	Name     string   `json:"name,omitempty"`
	Creator  string   `json:"creator,omitempty"`
	IsMember bool     `json:"is_member,omitempty"`
	Members  []string `json:"members,omitempty"`
}

// GetChannelInfoResp response of get group info
type GetChannelInfoResp struct {
	BaseResp
	Channel ChannelInfo `json:"channel"`
}

// GetChannelListResp response of get group info
type GetChannelListResp struct {
	BaseResp
	Channels []ChannelInfo `json:"groups,omitempty"`
}

// OpenChatResp response of open chat
type OpenChatResp struct {
	BaseResp
	Channel ChannelInfo `json:"channel,omitempty"`
}

// JoinChannelResp response of join channel
type JoinChannelResp BaseResp

// LeaveChannelResp response of leave channel
type LeaveChannelResp BaseResp

// CreateChannelResp response of create channel
type CreateChannelResp struct {
	BaseResp
	Chat struct {
		ChatID int64 `json:"chat_id"`
	} `json:"chat"`
}

// AddChannelMemberResp response of add channel member
type AddChannelMemberResp BaseResp

// DeleteChannelMemberResp response of delete channel member
type DeleteChannelMemberResp BaseResp

// PostNotificationResp response of PostNotification
type PostNotificationResp struct {
	Ok bool `json:"ok,omitempty"`
}
