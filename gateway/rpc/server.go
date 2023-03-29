package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"net/http"

	"github.com/rs/zerolog/log"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const InternalError = "internal server error"

const (
	Success            = 0
	ErrorCodeUnknow    = -101
	ErrorCodeInput     = -102
	ErrorCodeDelete    = -103
	ErrorCodeUnmarshal = -104
)

const (
	HealthCheckUrl = "/health"
	QuestionUrl    = "/api"
)

var SessionCookieName = "session_id"

var Host = "127.0.0.1"

const DefaultIndex = "terminus"
const DefaultMaxResult = 10

var once sync.Once

var client *http.Client

var RpcServer *Service

type Service struct {
	port     string
	zincUrl  string
	username string
	password string
}

func InitRpcService(url, port, username, password string) {
	once.Do(func() {
		client = &http.Client{Timeout: time.Minute * 3}
		RpcServer = &Service{
			port:     port,
			zincUrl:  url,
			username: username,
			password: password,
		}
	})
}

type LoggerMy struct {
}

func (*LoggerMy) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if strings.Index(msg, `"/healthcheck"`) > 0 {
		return
	}
	return
}

func (c *Service) Start(ctx context.Context) error {
	//start gin
	gin.DefaultWriter = &LoggerMy{}
	r := gin.Default()

	//cors middleware
	r.SetTrustedProxies(nil)
	r.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.POST("/api/input", c.HandleInput)
	r.POST("/api/delete", c.HandleDelete)
	r.POST("/api/query", c.HandleQuery)
	address := "0.0.0.0:" + c.port

	go r.Run(address)
	log.Info().Msgf("start rpc on port:%s", c.port)
	return nil
}

type Resp struct {
	ResultCode int    `json:"code"`
	ResultMsg  string `json:"data"`
}

func (s *Service) HandleInput(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		} else {
			c.JSON(http.StatusInternalServerError, rep)
		}
	}()

	index := c.PostForm("index")
	if index == "" {
		index = DefaultIndex
	}

	filename := c.PostForm("filename")

	content := c.PostForm("content")

	filePath := c.PostForm("path")

	size := int64(len([]byte(content)))

	fileHeader, err := c.FormFile("doc")
	if err == nil {
		file, err := fileHeader.Open()
		if err != nil {
			log.Error().Msgf("open file err %v", err)
			return
		}
		defer file.Close()
		docData, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error().Msgf("read file error %v", err)
			return
		}
		content = string(docData) //TODO:parse doc
		filename = fileHeader.Filename
		size = fileHeader.Size
	}

	if content == "" {
		log.Error().Msgf("content empty")
		rep.ResultMsg = "content empty"
		return
	}

	doc := map[string]interface{}{
		"name":        filename,
		"where":       filePath,
		"content":     content,
		"size":        size,
		"created":     time.Now().Unix(),
		"format_name": formatFilename(filename),
	}

	id, err := s.zincInput(index, doc)
	if err != nil {
		rep.ResultCode = ErrorCodeInput
		rep.ResultMsg = err.Error()
		return
	}
	rep.ResultCode = Success
	rep.ResultMsg = string(id)
	return
}

func (s *Service) HandleDelete(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		} else {
			c.JSON(http.StatusInternalServerError, rep)
		}
	}()

	index := c.PostForm("index")
	if index == "" {
		index = DefaultIndex
	}
	docId := c.PostForm("docId")
	if docId == "" {
		rep.ResultCode = ErrorCodeDelete
		rep.ResultMsg = fmt.Sprintf("docId empty")
		return
	}
	_, err := s.zincDelete(docId, index)
	if err != nil {
		rep.ResultCode = ErrorCodeDelete
		rep.ResultMsg = err.Error()
		return
	}
	rep.ResultCode = Success
	rep.ResultMsg = docId
}

type QueryResp struct {
	Count  int           `json:"count"`
	Offset int           `json:"offset"`
	Limit  int           `json:"limit"`
	Items  []QueryResult `json:"items"`
}

func (s *Service) HandleQuery(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		} else {
			c.JSON(http.StatusInternalServerError, rep)
		}
	}()

	index := c.PostForm("index")
	if index == "" {
		index = DefaultIndex
	}

	term := c.PostForm("query")
	if term == "" {
		rep.ResultMsg = "term empty"
		return
	}

	maxResults, err := strconv.Atoi(c.PostForm("limit"))
	if err != nil {
		maxResults = DefaultMaxResult
	}

	results, err := s.zincQuery(QueryReq{
		SearchType: "match",
		Query: Query{
			Term: term,
		},
		From:      0,
		MaxResult: maxResults,
	}, index)

	if err != nil {
		rep.ResultMsg = err.Error()
		return
	}

	rep.ResultCode = Success
	response := QueryResp{
		Count:  len(results),
		Offset: 0,
		Limit:  maxResults,
		Items:  results,
	}
	repMsg, _ := json.Marshal(&response)
	rep.ResultMsg = string(repMsg)
}
