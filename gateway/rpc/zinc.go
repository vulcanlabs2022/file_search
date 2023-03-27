package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	zinc "github.com/zinclabs/sdk-go-zincsearch"
)

// query = `{
// 	"search_type": "match",
// 	"query":
// 	{
// 		"term": "DEMTSCHENKO",
// 		"start_time": "2021-06-02T14:28:31.894Z",
// 		"end_time": "2021-12-02T15:28:31.894Z"
// 	},
// 	"from": 0,
// 	"max_results": 20,
// 	"_source": []
// }`

type QueryReq struct {
	SearchType string `json:"search_type"`
	Query      Query  `json:"query"`
	From       int    `json:"from"`
	MaxResult  int    `json:"max_results"`
}

type Query struct {
	Term string `json:"term"`
}

var ErrQuery = errors.New("query err")

func (s *Service) zincDelete(docId string, index string) ([]byte, error) {
	url := s.zincUrl + "/api/" + index + "/_doc/" + docId
	req, err := http.NewRequest("DELETE", url, strings.NewReader(""))
	if err != nil {
		return nil, ErrQuery
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrQuery
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ErrQuery
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrQuery
	}
	return body, nil
}

func (s *Service) zincInput(index string, document map[string]interface{}) ([]byte, error) {
	// string | Index
	id := uuid.NewString() // string | ID
	// document := map[string]interface{}{
	// 	"name":    "John Doe",
	// 	"age":     30,
	// 	"address": "123 Main St",
	// } // map[string]interface{} | Document

	ctx := context.WithValue(context.TODO(), zinc.ContextBasicAuth, zinc.BasicAuth{
		UserName: s.username,
		Password: s.password,
	})
	configuration := zinc.NewConfiguration()
	configuration.Servers = zinc.ServerConfigurations{
		zinc.ServerConfiguration{
			URL: s.zincUrl,
		},
	}

	apiClient := zinc.NewAPIClient(configuration)
	resp, _, err := apiClient.Document.IndexWithID(ctx, index, id).Document(document).Execute()
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Error when calling `Document.IndexWithID``: %v\n", err)
		// fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	// response from `IndexWithID`: MetaHTTPResponseID
	// fmt.Fprintf(os.Stdout, "Response from `Document.IndexWithID`: %v\n", resp.GetId())
	return []byte(resp.GetId()), nil
}

type Document struct {
	FilePath string `json:"filepath"`
	FileName string `json:"filename"`
	Content  string `json:"content"`
}

type QueryResult struct {
	Index    string `json:"index"`
	Where    string `json:"where"`
	Name     string `json:"name"`
	DocId    string `json:"docId"`
	Created  int64  `json:"time"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

func (s *Service) zincQuery(query QueryReq, index string) ([]QueryResult, error) {
	queryJson, _ := json.Marshal(&query)
	url := s.zincUrl + "/api/" + index + "/_search"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(queryJson)))
	if err != nil {
		return nil, ErrQuery
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrQuery
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ErrQuery
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrQuery
	}
	cnt := gjson.Get(string(body), "hits.total.value").Int()
	if cnt == 0 {
		return nil, errors.New(fmt.Sprintf("not found term %s", query.Query.Term))
	}
	results := make([]QueryResult, cnt)
	for i := 0; i < int(cnt); i++ {
		created := gjson.Get(string(body), fmt.Sprintf("hits.hits.%v._source.created", i)).Int()
		filename := gjson.Get(string(body), fmt.Sprintf("hits.hits.%v._source.name", i)).String()
		fileType := ""
		if nameSplits := strings.SplitN(filename, ".", 2); len(nameSplits) > 1 {
			fileType = nameSplits[1]
		}
		results[i] = QueryResult{
			Index:    index,
			Where:    gjson.Get(string(body), fmt.Sprintf("hits.hits.%v._source.where", i)).String(),
			Name:     filename,
			DocId:    gjson.Get(string(body), fmt.Sprintf("hits.hits.%v._id", i)).String(),
			Created:  created,
			Content:  gjson.Get(string(body), fmt.Sprintf("hits.hits.%v._source.content", i)).String(),
			Type:     fileType,
			Size:     gjson.Get(string(body), fmt.Sprintf("hits.hits.%v._source.size", i)).Int(),
			Modified: created,
		}
	}
	return results, nil
}
