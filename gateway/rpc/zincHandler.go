package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"wzinc/parser"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type QueryResp struct {
	Count  int           `json:"count"`
	Offset int           `json:"offset"`
	Limit  int           `json:"limit"`
	Items  []QueryResult `json:"items"`
}

func (s *Service) HandleInput(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		}
	}()

	index := c.PostForm("index")
	if index == "" {
		index = FileIndex
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
			rep.ResultMsg = err.Error()
			c.JSON(http.StatusBadRequest, rep)
			return
		}
		filename = fileHeader.Filename
		defer file.Close()
		content, err = parser.ParseDoc(file, filename)
		if err != nil {
			log.Error().Msgf("parse file error %v", err)
			rep.ResultMsg = err.Error()
			c.JSON(http.StatusBadRequest, rep)
			return
		}

		size = fileHeader.Size
	}

	if content == "" {
		log.Warn().Msgf("content empty")
	}

	doc := map[string]interface{}{
		"name":        filename,
		"where":       filePath,
		"content":     content,
		"size":        size,
		"created":     time.Now().Unix(),
		"updated":     time.Now().Unix(),
		"format_name": formatFilename(filename),
	}

	id, err := s.zincInput(index, doc)
	if err != nil {
		rep.ResultCode = ErrorCodeInput
		rep.ResultMsg = err.Error()
		c.JSON(http.StatusBadRequest, rep)
		return
	}
	rep.ResultCode = Success
	rep.ResultMsg = string(id)
}

func (s *Service) HandleDelete(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		}
	}()

	index := c.PostForm("index")
	if index == "" {
		index = FileIndex
	}
	docId := c.PostForm("docId")
	if docId == "" {
		rep.ResultCode = ErrorCodeDelete
		rep.ResultMsg = fmt.Sprintf("docId empty")
		c.JSON(http.StatusBadRequest, rep)
		return
	}
	_, err := s.zincDelete(docId, index)
	if err != nil {
		rep.ResultCode = ErrorCodeDelete
		rep.ResultMsg = err.Error()
		c.JSON(http.StatusBadRequest, rep)
		return
	}
	rep.ResultCode = Success
	rep.ResultMsg = docId
}

func (s *Service) HandleQuery(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}

	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		}
	}()

	index := c.PostForm("index")
	if index == "" {
		index = FileIndex
	}

	term := c.PostForm("query")
	if term == "" {
		rep.ResultMsg = "term empty"
		c.JSON(http.StatusBadRequest, rep)
		return
	}

	maxResults, err := strconv.Atoi(c.PostForm("limit"))
	if err != nil {
		maxResults = DefaultMaxResult
	}

	results, err := s.zincQuery(index, term)

	if err != nil {
		rep.ResultMsg = err.Error()
		c.JSON(http.StatusNotFound, rep)
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
