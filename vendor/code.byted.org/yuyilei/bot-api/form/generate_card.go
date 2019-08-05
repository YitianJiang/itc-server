package form

import (
	"fmt"
)

func GenerateCardMessageContent(card *CardForm) *MessageContent {
	return &MessageContent{Card: *card}
}

func GenerateCardLink(href string, pcHref *string, iosHref *string, androidHref *string) *CardElementForm {
	var cardLink CardElementForm
	cardLink.Href = href
	cardLink.PCHref = pcHref
	cardLink.IOSHref = iosHref
	cardLink.AndroidHref = androidHref
	return &cardLink
}

func isValidI18n(i18n *map[string]string) bool {
	if i18n == nil {
		return true
	}
	for key, _ := range *(i18n) {
		switch key {
			case "zh_ch",
			     "ja_jp",
			     "en_us": continue
		}
		return false
	}
	return true
}

func isValidImageColor(imageColor *string) bool {
	if imageColor == nil {
		return true
	}
	switch *(imageColor) {
		case "orange",
			 "red",
			 "yellow",
			 "gray",
			 "blue",
			 "green": return true
	}
	return false
}

func isValidMethod(method string) bool {
	switch method {
		case "jump",
			 "post",
			 "get": return true
	}
	return false
}

func GenerateCardHeader(title *string, i18n *map[string]string, imageColor *string, lines *int32) (*CardElementForm, error) {
	if !isValidI18n(i18n) {
		return nil, fmt.Errorf("i18n should in [zh_ch, ja_jp, en_us]")
	}
	if !isValidImageColor(imageColor) {
		return nil, fmt.Errorf("imagecolor should in [orange, red, yellow, gray, blue, green]")
	}
	var cardHeader CardElementForm
	cardHeader.Title = title
	if i18n != nil {
		cardHeader.I18n = *i18n
	}
	if imageColor != nil {
		cardHeader.ImageColor = *imageColor
	}
	cardHeader.Lines = lines
	return &cardHeader, nil
}

func GenerateButtonForm(text *string, i18n *map[string]string, triggeredText *string, triggeredI18n *map[string]string, method string,
	url string, needUserInfo bool, needMessageInfo bool, parameter *map[string]interface{}, openUrl *CardOpenUrlForm,
	hideOther *bool) (*CardButtonForm, error) {
	if text == nil {
		return nil, fmt.Errorf("text must required")
	}
	if !isValidI18n(i18n) {
		return nil, fmt.Errorf("i18n should in [zh_ch, ja_jp, en_us]")
	}
	if !isValidI18n(triggeredI18n) {
		return nil, fmt.Errorf("trigger_i18n should in [zh_ch, ja_jp, en_us]")
	}
	if !isValidMethod(method) {
		return nil, fmt.Errorf("method should in [post, get, jump]")
	}
	var buttonForm CardButtonForm
	buttonForm.Text = text
	buttonForm.TriggeredText = triggeredText
	buttonForm.Method = method
	buttonForm.Url = &url
	buttonForm.NeedUserInfo = &needUserInfo
	buttonForm.NeedMessageInfo = &needMessageInfo
	buttonForm.HideOthers = hideOther
	if parameter != nil {
		buttonForm.Parameter = *parameter
	}
	if openUrl != nil {
		buttonForm.OpenUrl = *openUrl
	}
	if i18n != nil {
		buttonForm.I18n = *i18n
	}
	if triggeredI18n != nil {
		buttonForm.TriggeredI18n = *triggeredI18n
	}
	return &buttonForm, nil
}

func GenerateCardForm(cardLink, header *CardElementForm, content [][]CardElementForm, actions []CardActionForm) *CardForm {
	var cardForm CardForm
	cardForm.CardLink = cardLink
	cardForm.Header = header
	cardForm.Content = content
	cardForm.Actions = actions
	return &cardForm
}

func GenerateOpenUrlForm(pcUrl string, iosUrl string, androidUrl string) *CardOpenUrlForm {
	var openUrlForm CardOpenUrlForm
	openUrlForm.PcUrl = &pcUrl
	openUrlForm.IosUrl =&iosUrl
	openUrlForm.AndroidUrl = &androidUrl
	return &openUrlForm
}