package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func UploadImageToCloudflare(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		os.Getenv("CLOUDFLARE_IMAGES_URL"),
		body,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CLOUDFLARE_IMAGES_TOKEN"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		log.Println(res.Status)
		return "", fmt.Errorf("unexpected status: %s", res.Status)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var responseObject struct {
		Result struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	json.Unmarshal(resBody, &responseObject)

	return responseObject.Result.ID, nil
}

func DeleteImageFromCloudflare(id string) error {
	req, err := http.NewRequest(
		"DELETE",
		os.Getenv("CLOUDFLARE_IMAGES_URL")+"/"+id,
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CLOUDFLARE_IMAGES_TOKEN"))

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
