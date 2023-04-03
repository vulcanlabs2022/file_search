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
	Count  int         `json:"count"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
	Items  []QueryItem `json:"items"`
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
		"format_name": FormatFilename(filename),
	}

	id, err := s.ZincInput(index, doc)
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
	_, err := s.ZincDelete(docId, index)
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
	items := slashQueryResult(results)
	response := QueryResp{
		Count:  len(items),
		Offset: 0,
		Limit:  maxResults,
		Items:  items,
	}
	repMsg, _ := json.Marshal(&response)
	rep.ResultMsg = string(repMsg)
}

type QueryItem struct {
	Index    string `json:"index"`
	Where    string `json:"where"`
	Name     string `json:"name"`
	DocId    string `json:"docId"`
	Created  int64  `json:"created"`
	Updated  int64  `json:"updated"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	Snippet  string `json:"snippet"`
}

func slashQueryResult(results []QueryResult) []QueryItem {
	type record struct {
		QueryItem
		id int
	}
	itemsMap := make(map[string]record)
	itemsList := make([]QueryItem, 0)
	id := 0
	for _, res := range results {
		if item, ok := itemsMap[res.Where]; ok {
			if res.Modified > item.Modified {
				shortRes := shortQueryResult(res)
				itemsMap[res.Where] = record{
					QueryItem: shortRes,
					id:        item.id,
				}
				itemsList[item.id] = shortRes
			}
			continue
		}
		shortRes := shortQueryResult(res)
		itemsList = append(itemsList, shortRes)
		itemsMap[res.Where] = record{
			QueryItem: shortQueryResult(res),
			id:        id,
		}
		id++
	}
	return itemsList
}

func shortQueryResult(res QueryResult) QueryItem {
	snippet := ""
	if len(res.HightLights) > 0 {
		snippet = res.HightLights[0]
	}
	return QueryItem{
		Index:    res.Index,
		Where:    res.Where,
		Name:     res.Name,
		DocId:    res.DocId,
		Created:  res.Created,
		Updated:  res.Updated,
		Type:     res.Type,
		Size:     res.Size,
		Modified: res.Modified,
		Snippet:  snippet,
	}
}
