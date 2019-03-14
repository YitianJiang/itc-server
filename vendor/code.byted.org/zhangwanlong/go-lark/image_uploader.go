package lark

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func postFormWithFile(req *http.Request) (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	return resp, err
}

func postFormWithFileRequest(url string, data map[string]string, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range data {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

// UploadImage to server
func UploadImage(token, path string) (*http.Response, error) {
	data := make(map[string]string)
	data["token"] = token
	req, err := postFormWithFileRequest(UploadImageURL, data, path)
	if err != nil {
		return nil, err
	}
	resp, err := postFormWithFile(req)
	return resp, err
}
