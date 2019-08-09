package form

import (
	"fmt"
)

func GenerateRichTextMessageContent(richText map[string]*PostForm) *MessageContent {
	return &MessageContent{Post: richText}
}

// 生成 Text 标签
func GenerateTextTag(text *string, unEscape bool, lines *int32)  *CardElementForm {
	var textTag CardElementForm
	textTag.Tag = "text"
	textTag.Text = text
	textTag.UnEscape = unEscape
	textTag.Lines = lines
	return &textTag
}

// 生成 A 标签
func GenerateATag(text *string, unEscape bool, href string) *CardElementForm {
	var aTag CardElementForm
	aTag.Tag = "a"
	aTag.Text = text
	aTag.UnEscape = unEscape
	aTag.Href = href
	return &aTag
}

// 生成 At 标签
func GenerateAtTag(text *string, userID string) *CardElementForm {
	var atTag CardElementForm
	atTag.Tag = "at"
	atTag.Text = text
	atTag.UserId = userID
	return &atTag
}

// 生成 Image 标签
func GenerateImageTag(imageKey string, height int32, width int32) *CardElementForm {
	var imageTag CardElementForm
	imageTag.Tag = "img"
	imageTag.ImageKey = imageKey
	imageTag.Height = height
	imageTag.Width = width
	return &imageTag
}

// 生成 field 标签
func GenerateFieldTag(fieldForm []CardFieldForm) (*CardElementForm, error){
	for _, eachField := range fieldForm {
		if eachField.Title == nil {
			return nil, fmt.Errorf("field's title must required")
		}
		if eachField.Title.Tag != "text" {
			return nil, fmt.Errorf("tag of field's title must be text")
		}
		if eachField.Value == nil {
			return nil, fmt.Errorf("field's value must required")
		}
		if eachField.Value.Tag != "text" {
			return nil, fmt.Errorf("tag of field's value must be text")
		}
		if eachField.Short == nil {
			return nil, fmt.Errorf("field's short must required")
		}
	}
	var fieldTag CardElementForm
	fieldTag.Tag = "field"
	fieldTag.Fields = fieldForm
	return &fieldTag, nil
}


func GenerateRichTextForm(title *string, content [][]CardElementForm) *PostForm {
	postForm := &PostForm{Title: title, Content:content}
	return postForm
}

