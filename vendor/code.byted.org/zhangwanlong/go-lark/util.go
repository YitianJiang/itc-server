package lark

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// PrintResponseBody prints response body
func PrintResponseBody(resp *http.Response) {
	buf, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(buf))
}

// BuildOutcomingMessageReq for msg builder
func BuildOutcomingMessageReq(token string, om OutcomingMessage) map[string]interface{} {
	params := map[string]interface{}{
		"token":   token,
		"chat_id": om.ChatID,
		"content": map[string]string{
			"image_key": om.Content.ImageKey,
			"text":      om.Content.Text,
			"title":     om.Content.Title,
			"chat_id":   om.Content.ChatID,
		},
		"mention_user_list": om.MetionUserIDs,
		"msg_type":          om.MsgType,
		"root_id":           om.RootID,
	}
	return params
}

// BuildPrivateMessageReq for msg builder
func BuildPrivateMessageReq(token string, om OutcomingMessage) map[string]interface{} {
	params := map[string]interface{}{
		"token":      token,
		"email":      om.Email,
		"chatter_id": om.ChatID,
		"content": map[string]string{
			"image_key": om.Content.ImageKey,
			"text":      om.Content.Text,
			"title":     om.Content.Title,
			"chat_id":   om.Content.ChatID,
		},
		"mention_user_list": om.MetionUserIDs,
		"msg_type":          om.MsgType,
		"root_id":           om.RootID,
	}
	return params
}

// DownloadFile downloads from a URL to local path
func DownloadFile(path, url string) error {
	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}
