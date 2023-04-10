package common

import (
	"context"
)

type Question struct {
	Message        string `json:"message"`
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
	Model          string `json:"model"`
	FilePath       string `json:"filepath"`
}

type QA struct {
	Question       Question `json:"question"`
	AnswerRole     string   `json:"answerRole"`
	Answer         string   `json:"answer"`
	MessageId      string   `json:"messageId"`
	ConversationId string   `json:"conversationId"`
	Model          string   `json:"model"`
}

type PendingQuestion struct {
	Data   Question
	Chunk  chan (RelayResponse)
	Finish chan (AnswerStreamFinish)
	Ctx    context.Context
}

type AnswerStreamFinish struct {
	Url            string `json:"url"` //bs url or openai key
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
	Model          string `json:"model"`
}

type RelayResponse struct {
	Url            string `json:"url"` //bs url or openai key
	Text           string `json:"text"`
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
	Model          string `json:"model"`
}

type CallbackResponse struct {
	Text           string `json:"text"`
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
	Model          string `json:"model"`
	Done           bool   `json:"done"`
}

type UserStatus struct {
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
	Url            string `json:"url"` //bs url
	LastTime       int64  `json:"lastTime"`
	Model          string `json:"model"`
}

type ProxyResponse struct {
	Text           string `json:"text"`
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
	Model          string `json:"model"`
}
