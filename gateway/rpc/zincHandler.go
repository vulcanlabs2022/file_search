package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
	"wzinc/common"
	"wzinc/parser"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type FileQueryResp struct {
	Count  int             `json:"count"`
	Offset int             `json:"offset"`
	Limit  int             `json:"limit"`
	Items  []FileQueryItem `json:"items"`
}

func (s *Service) HandleFileInput(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		}
	}()

	index := FileIndex

	filename := c.PostForm("filename")

	content := c.PostForm("content")

	filePath := c.PostForm("path")

	size := int64(len([]byte(content)))

	md5 := common.Md5File(bytes.NewReader([]byte(content)))

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
		data, _ := ioutil.ReadAll(file)
		file.Close()
		r := bytes.NewReader(data)
		content, err = parser.ParseDoc(r, filename)
		if err != nil {
			log.Error().Msgf("parse file error %v", err)
			rep.ResultMsg = err.Error()
			c.JSON(http.StatusBadRequest, rep)
			return
		}
		md5 = common.Md5File(r)
		size = fileHeader.Size
	}

	if content == "" {
		log.Warn().Msgf("content empty")
	}

	doc := map[string]interface{}{
		"name":        filename,
		"where":       filePath,
		"md5":         md5,
		"content":     content,
		"size":        size,
		"created":     time.Now().Unix(),
		"updated":     time.Now().Unix(),
		"format_name": FormatFilename(filename),
	}

	log.Info().Msgf("add input file index %s doc %v", index, doc)
	id, err := s.ZincInput(index, doc)
	if err != nil {
		rep.ResultCode = ErrorCodeInput
		rep.ResultMsg = err.Error()
		log.Error().Msgf("zinc input error %s", err.Error())
		c.JSON(http.StatusBadRequest, rep)
		return
	}
	rep.ResultCode = Success
	rep.ResultMsg = string(id)
}

func (s *Service) HandleFileDelete(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		}
	}()

	index := FileIndex
	docId := c.PostForm("docId")
	if docId == "" {
		rep.ResultCode = ErrorCodeDelete
		rep.ResultMsg = fmt.Sprintf("docId empty")
		c.JSON(http.StatusBadRequest, rep)
		return
	}
	log.Info().Msgf("zinc delete index %s docid%s", index, docId)
	_, err := s.ZincDelete(docId, index)
	if err != nil {
		rep.ResultCode = ErrorCodeDelete
		rep.ResultMsg = err.Error()
		log.Error().Msgf("zinc delete error %s", err.Error())
		c.JSON(http.StatusBadRequest, rep)
		return
	}
	rep.ResultCode = Success
	rep.ResultMsg = docId
}

func (s *Service) HandleFileQuery(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  "",
	}

	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		}
	}()

	index := FileIndex

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
	log.Info().Msgf("zinc query index %s term %s max %v", index, term, maxResults)
	results, err := s.zincQuery(index, term, int32(maxResults))

	if err != nil {
		rep.ResultMsg = err.Error()
		log.Error().Msg(rep.ResultMsg)
		c.JSON(http.StatusNotFound, rep)
		return
	}
	log.Debug().Msgf("zinc query results %v", results)

	rep.ResultCode = Success
	items := s.slashFileQueryResult(results)
	log.Debug().Msgf("zinc query items %v", items)
	response := FileQueryResp{
		Count:  len(items),
		Offset: 0,
		Limit:  maxResults,
		Items:  items,
	}
	repMsg, _ := json.Marshal(&response)
	rep.ResultMsg = string(repMsg)
	log.Debug().Msgf("response data %s", rep.ResultMsg)
}

type FileQueryItem struct {
	Index    string `json:"index"`
	Where    string `json:"where"`
	Name     string `json:"name"`
	DocId    string `json:"docId"`
	Created  int64  `json:"created"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	Snippet  string `json:"snippet"`
}

func (s *Service) slashFileQueryResult(results []FileQueryResult) []FileQueryItem {
	type record struct {
		FileQueryItem
		id int
	}
	itemsMap := make(map[string]record)
	itemsList := make([]FileQueryItem, 0)
	id := 0
	for _, res := range results {
		fileInfo, err := os.Stat(res.Where)
		if os.IsNotExist(err) {
			//delete if not exist
			log.Info().Msgf("zinc delete query found but not exist file %s id %s", res.Where, res.DocId)
			_, err := s.ZincDelete(res.DocId, FileIndex)
			if err != nil {
				log.Error().Msgf("zinc delete file error path %s id %s", res.Where, res.DocId)
			}
			continue
		}
		if err == nil {
			res.Size = fileInfo.Size()
		}
		if item, ok := itemsMap[res.Where]; ok {
			if res.Modified > item.Modified {
				shortRes := shortFileQueryResult(res)
				itemsMap[res.Where] = record{
					FileQueryItem: shortRes,
					id:            item.id,
				}
				itemsList[item.id] = shortRes
			}
			continue
		}
		shortRes := shortFileQueryResult(res)
		itemsList = append(itemsList, shortRes)
		itemsMap[res.Where] = record{
			FileQueryItem: shortFileQueryResult(res),
			id:            id,
		}
		id++
	}
	return itemsList
}

func shortFileQueryResult(res FileQueryResult) FileQueryItem {
	snippet := ""
	if len(res.HightLights) > 0 {
		snippet = res.HightLights[0]
	}
	return FileQueryItem{
		Index:    res.Index,
		Where:    res.Where,
		Name:     res.Name,
		DocId:    res.DocId,
		Created:  res.Created,
		Type:     res.Type,
		Size:     res.Size,
		Modified: res.Modified,
		Snippet:  snippet,
	}
}
