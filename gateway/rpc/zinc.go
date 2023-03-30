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

const ContentFieldName = "content"

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
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ErrQuery
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

type HightLight struct {
	Cnt     int    `json:"cnt"`
	Snippet string `json:"snippet"`
}

type QueryResult struct {
	Index       string       `json:"index"`
	Where       string       `json:"where"`
	Name        string       `json:"name"`
	DocId       string       `json:"docId"`
	Created     int64        `json:"time"`
	Content     string       `json:"content"`
	Type        string       `json:"type"`
	Size        int64        `json:"size"`
	Modified    int64        `json:"modified"`
	HightLights []HightLight `json:"highlight"`
}

func (s *Service) zincRawQuery(indexName, term string) (*zinc.MetaSearchResponse, error) {
	query := *zinc.NewMetaZincQuery()
	highlight := zinc.NewMetaHighlight()
	highlightContent := zinc.NewMetaHighlight()
	highlight.SetFields(map[string]zinc.MetaHighlight{"content": *highlightContent})
	query.SetHighlight(*highlight)

	matchQuery := *zinc.NewMetaMatchQuery()
	matchQuery.SetQuery(term)
	subQuery := *zinc.NewMetaQuery()
	subQuery.SetMatch(map[string]zinc.MetaMatchQuery{
		"content": matchQuery,
	})
	boolQuery := *zinc.NewMetaBoolQuery()
	boolQuery.SetShould([]zinc.MetaQuery{subQuery})
	queryQuery := *zinc.NewMetaQuery()
	queryQuery.SetBool(boolQuery)
	query.SetQuery(queryQuery)

	ctx := context.WithValue(context.Background(), zinc.ContextBasicAuth, zinc.BasicAuth{
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
	resp, _, err := apiClient.Search.Search(ctx, indexName).Query(query).Execute()
	if err != nil {
		return nil, fmt.Errorf("error when calling `SearchApi.Search``: %v", err)
	}
	return resp, nil
	// for _, hit := range resp.Hits.Hits {

	// 	for _, highlightRes := range hit.Highlight {
	// 		// fmt.Printf("hightlight %v\n", highlightRes)
	// 		for _, hh := range highlightRes.([]interface{}) {
	// 			fmt.Println(hh.(string))
	// 		}
	// 	}
	// }
}

func getFileQueryResult(resp *zinc.MetaSearchResponse) ([]QueryResult, error) {
	resultList := make([]QueryResult, 0)
	for _, hit := range resp.Hits.Hits {
		result := QueryResult{
			Index:       FileIndex,
			Where:       "",
			Name:        "",
			DocId:       "",
			Created:     0,
			Content:     "",
			Type:        "",
			Size:        0,
			Modified:    0,
			HightLights: []HightLight{},
		}
		if where, ok := hit.Fields["where"].(string); ok {
			result.Where = where
		}
		if name, ok := hit.Fields["name"].(string); ok {
			result.Name = name
		}
		result.DocId = *hit.Id
		if size, ok := hit.Fields["size"].(int64); ok {
			result.Size = size
		}

		for _, highlightRes := range hit.Highlight {
			// fmt.Printf("hightlight %v\n", highlightRes)
			for _, hh := range highlightRes.([]interface{}) {
				fmt.Println(hh.(string))

			}
		}
	}

}

func (s *Service) zincQuery(query QueryReq, index string) ([]QueryResult, error) {
	queryJson, _ := json.Marshal(&query)
	url := s.zincUrl + "/api/" + index + "/_search"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(queryJson)))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(string(body))
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

func (s *Service) listIndex() ([]string, error) {
	ctx := context.WithValue(context.Background(), zinc.ContextBasicAuth, zinc.BasicAuth{
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
	resp, r, err := apiClient.Index.IndexNameList(ctx).Execute()
	if err != nil {
		return nil, err
	}
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("Full HTTP response: %v\n", r)
	}
	return resp, nil
}

func (s *Service) createIndex(indexName string) error {
	index := *zinc.NewMetaIndexSimple()
	index.SetName(indexName)

	ctx := context.WithValue(context.Background(), zinc.ContextBasicAuth, zinc.BasicAuth{
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
	_, r, err := apiClient.Index.Create(ctx).Data(index).Execute()
	if err != nil {
		return err
	}
	// response from `Create`: MetaHTTPResponseIndex
	if r.StatusCode != 200 {
		e, _ := err.(*zinc.GenericOpenAPIError)
		me, _ := e.Model().(zinc.MetaHTTPResponseError)
		return fmt.Errorf("`Index.Create` error: %v", me.GetError())
	}
	return nil
}

func (s *Service) setupIndex() error {
	existIndexNameList, err := s.listIndex()
	if err != nil {
		return err
	}
	nameMap := make(map[string]bool)
	for _, existName := range existIndexNameList {
		nameMap[existName] = true
	}

	expectIndexList := []string{RssIndex, FileIndex}
	for _, indexName := range expectIndexList {
		if _, ok := nameMap[indexName]; !ok {
			err = s.createIndex(RssIndex)
			if err != nil {
				return err
			}
		}
		err = s.setIndexMapping(RssIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

//add highlightable filed "content" in index map setting
func (s *Service) setIndexMapping(indexName string) error {
	ctx := context.WithValue(context.Background(), zinc.ContextBasicAuth, zinc.BasicAuth{
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

	mapping := *zinc.NewMetaMappings() // MetaMappings | Mapping

	content := zinc.NewMetaProperty()
	content.SetType("text")
	content.SetIndex(true)
	content.SetHighlightable(true)

	mapping.SetProperties(map[string]zinc.MetaProperty{
		ContentFieldName: *content,
	})

	_, r, err := apiClient.Index.SetMapping(ctx, indexName).Mapping(mapping).Execute()
	if err != nil {
		return err
	}
	if r.StatusCode != 200 {
		e, _ := err.(*zinc.GenericOpenAPIError)
		me, _ := e.Model().(zinc.MetaHTTPResponseError)
		return fmt.Errorf("`Index.SetMapping` error: %v", me.GetError())
	}
	return nil
}
