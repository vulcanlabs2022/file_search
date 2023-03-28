package rpc

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/rs/zerolog/log"
)

func HttpPost(requrl, body string, timeoutS int, headerMap map[string]string) (payload []byte, err error) {
	req, err := http.NewRequest("POST", requrl, bytes.NewBufferString(body))
	if err != nil {
		log.Error().Msgf("Failed to new http request: %v", err.Error())
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

func formatFilename(filename string) string {
	formated := addSpaceBetweenCase(filename)
	formated = strings.ReplaceAll(formated, "_", " ")
	formated = strings.ReplaceAll(formated, "-", " ")
	formated = strings.ReplaceAll(formated, ".", " ")
	return formated
}

func addSpaceBetweenCase(str string) string {
	var result string
	for i, r := range str {
		if i > 0 {
			last := rune(str[i-1])
			if unicode.IsLower(last) && unicode.IsUpper(r) {
				result += " "
			}
		}
		result += string(r)
	}
	return result
}
