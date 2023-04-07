package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"
	"wzinc/common"
	"wzinc/db"
	"wzinc/parser"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const SensitiveResponse = "Sorry, as an artificial intelligence, I am unable to provide you with a standardized and satisfactory answer to your question. We will continuously optimize the system to provide better service. Thank you for your question."
const WaitForAIAnswer = time.Minute * 3

var userStatus UserStatus

func (s *Service) StartChatService(ctx context.Context) {
	//max concurrent questions
	handling := make(chan struct{}, s.maxPendingLength)
	for {
		select {
		case <-ctx.Done():
			return
		case qu := <-s.questionCh:
			handling <- struct{}{}
			log.Debug().Msgf("try send question to relay %s", qu.data)
			go func() {
				s.checkOneQuestion(qu)
				log.Debug().Msgf("question to relay done %s", qu.data)
				<-handling
			}()
		}
	}
}

func (s *Service) checkOneQuestion(qu pendingQuestion) {
	select {
	case <-qu.cancel:
		log.Warn().Msgf("close for timeout %v", qu.data)
		return
	default:
		defer close(qu.resp)
		// if _, ok := s.bsApiClient[qu.data.Model]; ok {
		// 	err := s.handleBsQuestion(qu)
		// 	if err != nil {
		// 		log.Error().Msgf("handle bs question error %s", err.Error())
		// 	}
		// 	return
		// }
		// log.Error().Msgf("model not exist name%s", qu.data.Model)
		s.handleFakeBsQuestion(qu)
		return
	}
}

func (s *Service) handleFakeBsQuestion(qu pendingQuestion) error {
	conversationId := qu.data.ConversationId
	if conversationId == "" {
		conversationId = uuid.New().String()
	}
	res := RelayResponse{
		Url:            "",
		Text:           "Hello, this is test",
		MessageId:      uuid.New().String(),
		ConversationId: conversationId,
		Model:          "test",
	}
	log.Info().Msgf("question: %s \n answer: %s \n model: %s", qu.data.Message, res.Text, res.Model)
	qu.resp <- res
	return nil
}

func (s *Service) handleBsQuestion(qu pendingQuestion) error {
	client, _ := s.bsApiClient[qu.data.Model]
	qa, err := client.GetAnswer(context.Background(), qu.data)
	if err != nil {
		return err
	}
	res := RelayResponse{
		Url:            client.Url,
		Text:           qa.Answer,
		MessageId:      qa.MessageId,
		ConversationId: qa.ConversationId,
		Model:          qu.data.Model,
	}
	log.Info().Msgf("question: %s \n answer: %s \n model: %s", qu.data.Message, res.Text, res.Model)
	qu.resp <- res
	return nil
}

func HandleRefresh(c *gin.Context) {
	defer func() {
		c.String(http.StatusOK, "success")
	}()
	userStatus = emptyStatus
}

func (s *Service) HandleQuestion(c *gin.Context) {
	statusCode := http.StatusBadRequest
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			statusCode = http.StatusOK
		}
		c.JSON(statusCode, rep)
	}()

	//get session state
	msg := c.PostForm("message")
	if msg == "" {
		rep.ResultMsg = "no question message"
		log.Error().Msg(rep.ResultMsg)
		return
	}
	modelName := c.PostForm("model")

	conv_id := c.PostForm("conversationId")
	filePath := c.PostForm("path")
	if filePath != "" {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			rep.ResultMsg = err.Error()
			log.Error().Msg(rep.ResultMsg)
			return
		}
		if !parser.IsParseAble(fileInfo.Name()) {
			rep.ResultMsg = "file not parsable"
			log.Error().Msg(rep.ResultMsg)
			return
		}
	}

	q := common.Question{
		Message:        msg,
		MessageId:      "",
		ConversationId: conv_id,
		Model:          modelName,
		FilePath:       filePath,
	}
	pendingQ := pendingQuestion{
		data:   q,
		resp:   make(chan RelayResponse, 1),
		cancel: make(chan struct{}),
	}
	defer close(pendingQ.cancel)

	timer := time.NewTimer(WaitForAIAnswer)

	select {
	case <-timer.C:
		log.Warn().Msgf("pending question time out %v", q)
		statusCode = http.StatusGatewayTimeout
		rep.ResultMsg = "AI model handle question timeout"
		return
	case s.questionCh <- pendingQ:
		select {
		case <-timer.C:
			log.Warn().Msgf("pending question time out %v", q)
			statusCode = http.StatusGatewayTimeout
			rep.ResultMsg = "AI model handle question timeout"
			return
		case answer := <-pendingQ.resp:
			if answer.Text == "" {
				statusCode = http.StatusGatewayTimeout
				rep.ResultMsg = "AI model handle question error"
				return
			}

			go func() {
				err := db.InsertSingleConversation(db.Message{
					ConversationId: answer.ConversationId,
					MessageId:      answer.MessageId,
					Prompt:         q.Message,
					Text:           answer.Text,
					StartTime:      time.Now().Unix(),
					Model:          answer.Model,
				})
				if err != nil {
					log.Error().Msgf("insert into db error %s", err.Error())
				}
			}()

			//update user status
			userStatus = UserStatus{
				MessageId:      answer.MessageId,
				ConversationId: answer.ConversationId,
				Url:            answer.Url,
				LastTime:       time.Now().Unix(),
				Model:          answer.Model,
			}

			var data []byte
			data, _ = json.Marshal(&ProxyResponse{
				Text:           answer.Text,
				MessageId:      answer.MessageId,
				ConversationId: answer.ConversationId,
				Model:          answer.Model,
			})
			rep.ResultMsg = string(data)
			rep.ResultCode = Success
			return
		}
	}
}
