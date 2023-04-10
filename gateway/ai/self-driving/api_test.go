package selfdriving

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
	"wzinc/common"
)

func TestGetAnswerFile(t *testing.T) {
	resp, err := common.HttpPostFile("", MaxPostTimeOut, map[string]string{
		common.PostFileParamKey:  "",
		common.PostQueryParamKey: "who is elon mask?",
	})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	tickers := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-tickers.C:
			fmt.Println("try reading...")
			s, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				panic(err)
			}
			fmt.Println(s)
		}
	}
}

func TestSse(t *testing.T) {
	url := "http://localhost:8000/stream"
	client := http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	for {
		event, err := readEvent(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Event: %v\n", event)
	}
}

func readEvent(body io.Reader) (string, error) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		if event := scanner.Text(); event != "" {
			return event, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func TestScanner(t *testing.T) {
	delimStr := "}\n"
	delim := []byte(delimStr)
	data := "aaa sdfd fa\n df k" + delimStr + "adf" + delimStr + "adf" + delimStr

	scanner := bufio.NewScanner(strings.NewReader(data))
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, delim); i >= 0 {
			return i + len(delim), append(data[:i], byte('}')), nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	})

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
