package lark

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// LogPrefix for lark
const LogPrefix = "[go-lark] "

const (
	// ChatBot should call NewChatBot or NewBot
	// Create from https://lark-open.bytedance.net/list
	ChatBot = iota
	// NotificationBot for webhook, behave as a simpler notification bot
	// Create from Lark group
	// https://lark-open.bytedance.net/doc/fuHCWYbPdHZTGODh1DbiIa#jjJE6r
	NotificationBot
	// H5App TODO
	// https://docs.bytedance.net/doc/qzKSTalt6UST5DDa6YWI5g
	H5App
)

// Bot lark.Bot
type Bot struct {
	botType int
	token   string
	webhook string
	client  *http.Client
}

// NewBot is an alias of NewChatBot
func NewBot(token string) *Bot {
	return NewChatBot(token)
}

// NewChatBot init a bot with token
func NewChatBot(token string) *Bot {
	log.SetPrefix(LogPrefix)
	return &Bot{
		botType: ChatBot,
		token:   token,
		client:  initClient(),
	}
}

// NewNotificationBot with URL
func NewNotificationBot(hookURL string) *Bot {
	log.SetPrefix(LogPrefix)
	return &Bot{
		botType: NotificationBot,
		webhook: hookURL,
		client:  initClient(),
	}
}

func initClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}

// httpPost send http posts
func (bot *Bot) httpPost(url string, params map[string]interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(params)
	if err != nil {
		log.Printf("Encode json failed: %+v\n", err)
		return nil, err
	}
	resp, err := bot.client.Post(url, "application/json; charset=utf-8", buf)
	return resp, err
}

// requireType checks whether the action is allowed in a list of bot types
func (bot *Bot) requireType(botType ...int) bool {
	for _, iterType := range botType {
		if bot.botType == iterType {
			return true
		}
	}
	return false
}

// SetClient assigns a new client to bot.client
func (bot *Bot) SetClient(c *http.Client) {
	bot.client = c
}

// PostNotification send message to a webhook
func (bot *Bot) PostNotification(title, text string) (*PostNotificationResp, error) {
	if !bot.requireType(NotificationBot) {
		return nil, errors.New("Bot type error")
	}
	params := map[string]interface{}{
		"title": title,
		"text":  text,
	}
	resp, err := bot.httpPost(bot.webhook, params)
	if err != nil {
		log.Printf("PostNotification failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData PostNotificationResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	return &respData, err
}

// PostText to channel
func (bot *Bot) PostText(text, channel string) (*PostMessageResp, error) {
	mb := NewMsgBuffer(MsgText)
	message := mb.BindChannel(channel).Text(text).Build()
	respData, err := bot.PostMessage(message)
	if err != nil {
		log.Printf("PostText failed: %+v\n", err)
		return nil, err
	}
	return respData, err
}

// PostContent to channel
func (bot *Bot) PostContent(title, content, channel string) (*PostMessageResp, error) {
	mb := NewMsgBuffer(MsgPost)
	message := mb.BindChannel(channel).Title(title).Text(content).Build()
	respData, err := bot.PostMessage(message)
	if err != nil {
		log.Printf("PostContent failed: %+v\n", err)
		return nil, err
	}
	return respData, err
}

// PostTextMention @somebody text
func (bot *Bot) PostTextMention(userID, text, channel string) (*PostMessageResp, error) {
	mb := NewMsgBuffer(MsgText)
	message := mb.BindChannel(channel).BindMention([]string{userID}).Text(text).Build()
	respData, err := bot.PostMessage(message)
	if err != nil {
		log.Printf("PostTextMention failed: %+v\n", err)
		return nil, err
	}
	return respData, err
}

// PostTextMentionAll @all
// Be cautious to use this func, because everyone will be bothered
func (bot *Bot) PostTextMentionAll(text, channel string) (*PostMessageResp, error) {
	mb := NewMsgBuffer(MsgText)
	message := mb.BindChannel(channel).MentionAll("All").Space(1).Text(text).Build()
	respData, err := bot.PostMessage(message)
	if err != nil {
		log.Printf("PostTextMentionAll failed: %+v\n", err)
		return nil, err
	}
	return respData, err
}

// PostShareChatCard sends share chat card
func (bot *Bot) PostShareChatCard(chatID, channel string) (*PostMessageResp, error) {
	mb := NewMsgBuffer(MsgShareCard)
	message := mb.BindChannel(channel).ShareChat(chatID).Build()
	respData, err := bot.PostMessage(message)
	if err != nil {
		log.Printf("PostShareChatCard failed: %+v\n", err)
		return nil, err
	}
	return respData, err
}

// PostImage sends an image
func (bot *Bot) PostImage(imageKey, channel string) (*PostMessageResp, error) {
	mb := NewMsgBuffer(MsgImage)
	message := mb.BindChannel(channel).Image(imageKey).Build()
	respData, err := bot.PostMessage(message)
	if err != nil {
		log.Printf("PostImage failed: %+v\n", err)
		return nil, err
	}
	return respData, err
}

// PostMessage allows user to construct messsage body and post it
func (bot *Bot) PostMessage(message OutcomingMessage) (*PostMessageResp, error) {
	resp, err := bot.httpPost(MessageURL, BuildOutcomingMessageReq(bot.token, message))
	if err != nil {
		log.Printf("PostMessage failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData PostMessageResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("PostMessage decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// PostPrivateMessage send private message
func (bot *Bot) PostPrivateMessage(message OutcomingMessage) (*PostPrivateMessageResp, error) {
	resp, err := bot.httpPost(PrivateChatURL, BuildPrivateMessageReq(bot.token, message))
	if err != nil {
		log.Printf("PostPrivateMessage failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData PostPrivateMessageResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("PostPrivateMessage decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// GetUserInfoByEmail get user info by email
func (bot *Bot) GetUserInfoByEmail(email string) (*UserInfoByEmailResp, error) {
	params := map[string]interface{}{
		"token": bot.token,
		"email": email,
	}
	resp, err := bot.httpPost(UserInfoByEmailURL, params)
	if err != nil {
		log.Printf("GetUserInfoByEmail failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData UserInfoByEmailResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("GetUserInfoByEmail decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// GetUserInfo get user info by user id
func (bot *Bot) GetUserInfo(userID string) (*UserInfoResp, error) {
	params := map[string]interface{}{
		"token": bot.token,
		"user":  userID,
	}
	resp, err := bot.httpPost(UserInfoURL, params)
	if err != nil {
		log.Printf("GetUserInfo failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData UserInfoResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("GetUserInfo decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// GetBotList returns bot list
// It can be called without bot token
func (bot *Bot) GetBotList(userToken string) (*BotListResp, error) {
	params := map[string]interface{}{
		"token": userToken,
	}
	resp, err := bot.httpPost(BotListURL, params)
	if err != nil {
		log.Printf("GetBotList failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData BotListResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("GetBotList decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// GetBotInfo returns bot info
// It can be called without bot token
func (bot *Bot) GetBotInfo(userToken string) (*BotInfoResp, error) {
	params := map[string]interface{}{
		"bot":   bot.token,
		"token": userToken,
	}
	resp, err := bot.httpPost(BotInfoURL, params)
	if err != nil {
		log.Printf("GetBotInfo failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData BotInfoResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("GetBotInfo failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// AddBotEvent create event for a bot
func (bot *Bot) AddBotEvent(userToken string, eventType int) (*AddEventResp, error) {
	err := deprecateFunc(bot.AddBotEvent, "AddBotEvent is only available only at <https://lark-open.bytedance.net/page/app>.")
	return nil, err
}

// AddBotEventHook create event for a bot
func (bot *Bot) AddBotEventHook(userToken, address, description string) (*AddEventHookResp, error) {
	err := deprecateFunc(bot.AddBotEventHook, "AddBotEventHook is only available only at <https://lark-open.bytedance.net/page/app>.")
	return nil, err
}

// UploadImage to server
func (bot *Bot) UploadImage(path string) (*UploadImageResp, error) {
	resp, err := UploadImage(bot.token, path)
	if err != nil {
		log.Printf("UploadImage failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData UploadImageResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("UploadImage decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// GetChannelInfo returns channel info
func (bot *Bot) GetChannelInfo(channel string) (*GetChannelInfoResp, error) {
	params := map[string]interface{}{
		"token":   bot.token,
		"channel": channel,
	}
	resp, err := bot.httpPost(ChannelInfoURL, params)
	if err != nil {
		log.Printf("GetChannelInfo failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData GetChannelInfoResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("GetChannelInfo decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// GetChannelList returns channel list of a user
func (bot *Bot) GetChannelList(userToken string) (*GetChannelListResp, error) {
	params := map[string]interface{}{
		"token": userToken,
	}
	resp, err := bot.httpPost(ChannelListURL, params)
	if err != nil {
		log.Printf("GetChannelList failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData GetChannelListResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("GetChannelList decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// OpenChat opens a chat with user
func (bot *Bot) OpenChat(userID string) (*OpenChatResp, error) {
	params := map[string]interface{}{
		"token": bot.token,
		"user":  userID,
	}
	resp, err := bot.httpPost(OpenChatURL, params)
	if err != nil {
		log.Printf("OpenChat failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData OpenChatResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("OpenChat decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// JoinChannel joins a chat channel
func (bot *Bot) JoinChannel(userToken, chatID string) (*JoinChannelResp, error) {
	params := map[string]interface{}{
		"bot":     bot.token,
		"token":   userToken,
		"chat_id": chatID,
	}
	resp, err := bot.httpPost(JoinChannelURL, params)
	if err != nil {
		log.Printf("JoinChannel failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData JoinChannelResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("JoinChannel decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// LeaveChannel leaves a chat channel
func (bot *Bot) LeaveChannel(userToken, chatID string) (*LeaveChannelResp, error) {
	params := map[string]interface{}{
		"bot":     bot.token,
		"token":   userToken,
		"chat_id": chatID,
	}
	resp, err := bot.httpPost(LeaveChannelURL, params)
	if err != nil {
		log.Printf("LeaveChannel failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData LeaveChannelResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("LeaveChannel decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// CreateChannel creates a chat channel
func (bot *Bot) CreateChannel(name, description, iconKey string, userIDs []string) (*CreateChannelResp, error) {
	params := map[string]interface{}{
		"token":       bot.token,
		"name":        name,
		"description": description,
		"icon_key":    iconKey,
		"user_ids":    userIDs,
	}
	resp, err := bot.httpPost(CreateChannelURL, params)
	if err != nil {
		log.Printf("CreateChannel failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData CreateChannelResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("CreateChannel decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// AddChannelMember delete a member of a channel
func (bot *Bot) AddChannelMember(chatID string, userIDs []string) (*AddChannelMemberResp, error) {
	params := map[string]interface{}{
		"token":    bot.token,
		"chat_id":  chatID,
		"user_ids": userIDs,
	}
	resp, err := bot.httpPost(AddChannelMemberURL, params)
	if err != nil {
		log.Printf("AddChannelMember failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData AddChannelMemberResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("AddChannelMember decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}

// DeleteChannelMember delete a member of a channel
func (bot *Bot) DeleteChannelMember(chatID string, userIDs []string) (*DeleteChannelMemberResp, error) {
	params := map[string]interface{}{
		"token":    bot.token,
		"chat_id":  chatID,
		"user_ids": userIDs,
	}
	resp, err := bot.httpPost(DeleteChannelMemberURL, params)
	if err != nil {
		log.Printf("DeleteChannelMember failed: %+v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	var respData DeleteChannelMemberResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		log.Printf("DeleteChannelMember decode body failed: %+v\n", err)
		return nil, err
	}
	return &respData, err
}
