package common

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

const PostFileParamKey = "file"
const PostQueryParamKey = "query"

func HttpPostFile(requrl string, timeoutS int, params map[string]string) (payload []byte, err error) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	for key, value := range params {
		if key == PostFileParamKey {
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
		return
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	client := &http.Client{Timeout: time.Second * time.Duration(timeoutS)}

	resp, err := client.Do(req)
	if resp == nil {
		return
	}

	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	return
}

func HttpPost(requrl, body string, timeoutS int, headerMap map[string]string) (payload []byte, err error) {
	req, err := http.NewRequest("POST", requrl, bytes.NewBufferString(body))
	if err != nil {
		log.Error().Msgf("Failed to new http request:%s", err.Error())
		return
	}

	for k, v := range headerMap {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: time.Second * time.Duration(timeoutS)}

	resp, err := client.Do(req)
	if resp == nil {
		return
	}

	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	return
}
