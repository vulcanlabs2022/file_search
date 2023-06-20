package selfdriving

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"wzinc/common"
	"wzinc/db"
	"wzinc/parser"

	"github.com/rs/zerolog/log"
)

const MaxMsgLogLength = 50
const MaxPromtLength = 20480
const MaxPostTimeOut = 300
const ReadTickerTime = time.Millisecond * 200
const JsonSuffix = "}\n"
const MaxTry = 1
const RetryWaitTime = time.Second

const FakeAnswer = "Elon Musk is an American entrepreneur, engineer and inventor who has founded several successful companies including SpaceX, Tesla Motors and SolarCity. He is also known for his work in developing electric cars and solar energy systems."

var MaxConversactionSuspend = 60 * 60 //1 hour

type BSRequest struct {
	Query   string     `json:"query"`
	History [][]string `json:"history"`
	Text    string     `json:"text"`
	Type    string     `json:"type"` //basic, single_doc, full_doc
}

type BSResponse struct {
	ErrCode  int    `json:"errcode"`
	Response string `json:"response"`
	Ret      int    `json:"ret"`
}

type Client struct {
	Url           string
	logUpdateTime map[string]int64
	ModelName     string
}

func NewClient(url, modelName string, ctx context.Context) *Client {
	c := &Client{
		Url:           url,
		logUpdateTime: make(map[string]int64),
		ModelName:     modelName,
	}
	return c
}

func (c *Client) buildPromt(q *common.Question) (*BSRequest, error) {
	maxLength := MaxPromtLength
	promtHistoryLen := 0
	promt := BSRequest{
		Query:   q.Message,
		History: [][]string{},
		Type:    q.Type,
	}
	conversationFrom := time.Now().Unix() - int64(MaxConversactionSuspend)
	msgLog, err := db.GetResentConversation(q.ConversationId, conversationFrom)
	if err != nil {
		log.Error().Msgf("GetResentConversation conversationid %s from timestamp %v error %s", q.ConversationId, conversationFrom, err.Error())
	}
	if err == nil {
		for i := len(msgLog) - 1; i >= 0; i-- {
			msg := msgLog[i]
			promt.History = append(promt.History, []string{msg.Prompt, msg.Text})
			promtHistoryLen = promtHistoryLen + len(msg.Prompt) + len(msg.Text)
			//pop early msgs if promt over length
			for promtHistoryLen > maxLength {
				if len(promt.History) <= 1 {
					break
				}
				shortLen := len(promt.History[0][0]) + len(promt.History[0][1])
				promtHistoryLen = promtHistoryLen - shortLen
				promt.History = promt.History[1:]
			}
		}
	}

	//parse doc
	if q.FilePath != "" {
		f, err := os.Open(q.FilePath)
		if err != nil {
			return nil, err
		}
		data, _ := ioutil.ReadAll(f)
		f.Close()
		r := bytes.NewReader(data)
		fileStr, err := parser.ParseDoc(r, q.FilePath)
		if err != nil {
			return nil, err
		}
		promt.Text = fileStr
	}
	return &promt, nil
}

func (c *Client) GetAnswerFake(ctx context.Context, qu common.PendingQuestion) (err error) {
	prompt, err := c.buildPromt(&qu.Data)
	if err != nil {
		return err
	}
	history, err := json.Marshal(prompt.History)
	if err != nil {
		return err
	}
	log.Debug().Msgf("history:%s", string(history))
	defer func() {
		qu.Finish <- common.AnswerStreamFinish{
			Url:            c.Url,
			MessageId:      qu.Data.MessageId,
			ConversationId: qu.Data.ConversationId,
			Model:          c.ModelName,
		}
	}()
	totalAnwer := ""
	fakelist := strings.Split(FakeAnswer, " ")
	time.Sleep(time.Second * 10)
	for _, word := range fakelist {
		time.Sleep(ReadTickerTime)
		totalAnwer = totalAnwer + " " + word
		qu.Chunk <- common.RelayResponse{
			Url:            c.Url,
			Text:           totalAnwer,
			MessageId:      qu.Data.MessageId,
			ConversationId: qu.Data.ConversationId,
			Model:          c.ModelName,
		}
	}
	return err
}

func (c *Client) GetAnswer(ctx context.Context, qu common.PendingQuestion) (err error) {
	defer func() {
		finish := common.AnswerStreamFinish{
			Url:            c.Url,
			MessageId:      qu.Data.MessageId,
			ConversationId: qu.Data.ConversationId,
			Model:          c.ModelName,
		}
		if err != nil {
			finish.ErrorMsg = err.Error()
		}
		qu.Finish <- finish
	}()
	prompt, err := c.buildPromt(&qu.Data)
	if err != nil {
		return err
	}
	history, err := json.Marshal(prompt.History)
	if err != nil {
		return err
	}
	log.Debug().Msgf("history:%s", string(history))

	body, err := json.Marshal(prompt)
	if err != nil {
		return err
	}
	log.Debug().Msgf("request body %s", string(body))
	tryTimes := 0
	tryPost := func() error {
		resp, err := common.HttpPost(c.Url, string(body), MaxPostTimeOut)
		if err != nil {
			log.Error().Msgf("post url %s body %v err %s", c.Url, string(body), err.Error())
			return err
		}
		err = c.pipeResponse(resp, qu)
		return err
	}

	for tryTimes < MaxTry {
		tryTimes++
		err = tryPost()
		if err != nil {
			time.Sleep(RetryWaitTime * time.Duration(tryTimes))
			continue
		} else {
			break
		}
	}
	return err
}

func (c *Client) pipeResponse(resp *http.Response, qu common.PendingQuestion) error {
	defer resp.Body.Close()
	reader := bufio.NewReaderSize(resp.Body, 40960)
	tic := time.NewTicker(ReadTickerTime)
	wordsJson := ""
	for {
		select {
		case <-qu.Ctx.Done():
			return nil
		case <-tic.C:
			words, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if wordsJson != "" {
						return errors.New("unexpected connection closed")
					} else {
						return nil
					}
				} else {
					log.Error().Msgf("read string from body err %s", err.Error())
					return err
				}
			}
			if words == "" {
				continue
			}
			log.Debug().Msgf("new words:%s", words)
			wordsJson = wordsJson + words
			if !strings.HasSuffix(wordsJson, JsonSuffix) {
				log.Debug().Msg("no suffix read more")
				continue
			}
			var bsResp BSResponse
			err = json.Unmarshal([]byte(wordsJson), &bsResp)
			if err != nil {
				log.Error().Msgf("pipe response unmarshal json err %s %s", wordsJson, err.Error())
				continue
			}
			wordsJson = ""
			log.Debug().Msgf("new chunk: %s", bsResp.Response)
			qu.Chunk <- common.RelayResponse{
				Url:            c.Url,
				Text:           bsResp.Response,
				MessageId:      qu.Data.MessageId,
				ConversationId: qu.Data.ConversationId,
				Model:          c.ModelName,
			}
		}
	}
}

func splitJson(data []byte, atEOF bool) (int, []byte, error) {
	delim := []byte(JsonSuffix)
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
}
