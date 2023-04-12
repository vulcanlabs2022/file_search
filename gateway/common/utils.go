package common

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

const PostFileParamKey = "file"
const PostQueryParamKey = "query"
const PostHistoryParamKey = "history"

func HttpPostFile(requrl string, timeoutS int, params map[string]string) (*http.Response, error) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	for key, value := range params {
		if key == PostFileParamKey {
			if value == "" {
				continue
			}
			file, err := os.Open(value)
			if err != nil {
				return nil, err
			}
			defer file.Close()
			part, err := multipartWriter.CreateFormFile("file", filepath.Base(file.Name()))
			if err != nil {
				return nil, err
			}
			if _, err = io.Copy(part, file); err != nil {
				return nil, err
			}
			continue
		}
		multipartWriter.WriteField(key, value)
	}
	multipartWriter.Close()

	req, err := http.NewRequest("POST", requrl, &requestBody)
	if err != nil {
		log.Error().Msgf("Failed to new http request:%s", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{Timeout: time.Second * time.Duration(timeoutS)}

	return client.Do(req)
}

func HttpPost(requrl, body string, timeoutS int) (*http.Response, error) {
	req, err := http.NewRequest("POST", requrl, bytes.NewBufferString(body))
	if err != nil {
		log.Error().Msgf("Failed to new http request:%s", err.Error())
		return nil, err
	}

	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * time.Duration(timeoutS)}

	resp, err := client.Do(req)
	return resp, err
}
