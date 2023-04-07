package selfdriving

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wzinc/common"
	"wzinc/db"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const MaxMsgLogLength = 50
const MaxPromtLength = 20480
const MaxPostTimeOut = 300

var MaxConversactionSuspend = 60 * 60 //1 hour

type BSRequest struct {
	Content string     `json:"content"`
	History [][]string `json:"history"`
	Model   string     `json:"model,omitempty"`
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

func (c *Client) buildPromt(q *common.Question) BSRequest {
	maxLength := MaxPromtLength
	promtHistoryLen := 0
	promt := BSRequest{
		Content: q.Message,
		History: [][]string{},
		//TODO: add model name
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
	return promt
}

func (c *Client) GetAnswer(ctx context.Context, q common.Question) (*common.QA, error) {
	if q.FilePath == "" {
		return c.getAnswerWorld(ctx, q)
	}
	return c.getAnswerFile(ctx, q)
}

func (c *Client) getAnswerWorld(ctx context.Context, q common.Question) (*common.QA, error) {
	prompt := c.buildPromt(&q)
	log.Debug().Msg("self driving model" + q.Model + "prompt:\n" + fmt.Sprintf("%v", prompt))
	promptData, err := json.Marshal(&prompt)
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}
	resp, err := common.HttpPost(c.Url, string(promptData), MaxPostTimeOut, map[string]string{})
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}
	var bsResp BSResponse
	err = json.Unmarshal(resp, &bsResp)
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}
	if q.ConversationId == "" {
		q.ConversationId = uuid.New().String()
	}
	qa := common.QA{
		Question:       q,
		AnswerRole:     "assitant",
		Answer:         bsResp.Response,
		MessageId:      uuid.NewString(),
		ConversationId: q.ConversationId,
		Model:          c.ModelName,
	}
	return &qa, nil
}

func (c *Client) getAnswerFile(ctx context.Context, q common.Question) (*common.QA, error) {
	resp, err := common.HttpPostFile(c.Url, MaxPostTimeOut, map[string]string{
		common.PostFileParamKey:  q.FilePath,
		common.PostQueryParamKey: q.Message,
	})
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}
	var bsResp BSResponse
	err = json.Unmarshal(resp, &bsResp)
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}
	if q.ConversationId == "" {
		q.ConversationId = uuid.New().String()
	}
	qa := common.QA{
		Question:       q,
		AnswerRole:     "assitant",
		Answer:         bsResp.Response,
		MessageId:      uuid.NewString(),
		ConversationId: q.ConversationId,
		Model:          c.ModelName,
	}
	return &qa, nil
}
