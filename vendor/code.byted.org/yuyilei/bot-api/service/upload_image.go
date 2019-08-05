package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

func GenerateBinaryImage(path string) (*bytes.Buffer, *string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open file fail, %v", err)
	}
	defer file.Close()
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	imageFile, err := writer.CreateFormFile("image", path)
	if err != nil {
		return nil, nil, fmt.Errorf("create form file fail, %v", err)
	}
	_, err = io.Copy(imageFile, file)
	if err != nil {
		return nil, nil, fmt.Errorf("copy file fail, %s", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("close image file fail, %s", err)
	}
	contentType := writer.FormDataContentType()
	return buffer, &contentType, nil
}

func (abot *BotService) UploadImage(path string) (map[string]interface{},error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("UploadImage failed because of can not get validTenantAccessToken, error= %v", err)
	}
	body, contentType, err := GenerateBinaryImage(path)
	if err != nil {
		return nil, fmt.Errorf("UploadImage failed because of GenerateBinaryImage fail, error= %v", err)
	}
	authorization := fmt.Sprintf("Bearer %s", abot.TenantAccessToken.Token)
	header := map[string]string{"Authorization":authorization, "Content-Type":*contentType }
	response, err := PostRequestForm(UploadImageUrl, header, body)
	if err != nil {
		return nil, fmt.Errorf("UploadImage failed because of sendRequest fail, error= %v", err)
	}
	return response, err
}