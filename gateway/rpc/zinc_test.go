package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	zinc "github.com/zinclabs/sdk-go-zincsearch"
)

const zincUrl = "http://localhost:4080"
const port = "1234"
const username = "admin"
const password = "User#123"

const index = DefaultIndex

func TestClientSearchV1(t *testing.T) { // string | Index
	query := *zinc.NewV1ZincQuery() // V1ZincQuery | Query
	query.SetSearchType("match")
	params := *zinc.NewV1QueryParams()
	params.SetTerm("123abc")
	query.SetQuery(params)
	query.SetSortFields([]string{"-@timestamp"})
	query.SetMaxResults(5)
	query.SetSource([]string{"content", "filename", "filepath"})

	ctx := context.WithValue(context.TODO(), zinc.ContextBasicAuth, zinc.BasicAuth{
		UserName: username,
		Password: password,
	})

	configuration := zinc.NewConfiguration()
	configuration.Servers = zinc.ServerConfigurations{
		zinc.ServerConfiguration{
			URL: zincUrl,
		},
	}
	apiClient := zinc.NewAPIClient(configuration)
	resp, r, err := apiClient.Search.SearchV1(ctx, index).Query(query).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SearchApi.SearchV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchV1`: V1SearchResponse
	fmt.Fprintf(os.Stdout, "Response from `SearchApi.SearchV1`: %v\n", resp)
	for _, hit := range resp.Hits.Hits {
		fmt.Printf("%v %v\n", hit.GetTimestamp(), hit.GetSource())
	}
}

func TestClientSearchV2(t *testing.T) {
	// eg:
	// {
	// 	"query": {
	// 		"bool": {
	// 			"should": [
	// 				{
	// 					"match": {
	// 						"name": {
	// 							"query": "John"
	// 						}
	// 					}
	// 				}
	// 			]
	// 		}
	// 	}
	// }
	query := *zinc.NewMetaZincQuery() // V1ZincQuery | Query
	matchQuery := *zinc.NewMetaMatchQuery()
	matchQuery.SetQuery("Download")
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
		UserName: username,
		Password: password,
	})

	configuration := zinc.NewConfiguration()
	configuration.Servers = zinc.ServerConfigurations{
		zinc.ServerConfiguration{
			URL: zincUrl,
		},
	}

	apiClient := zinc.NewAPIClient(configuration)
	resp, r, err := apiClient.Search.Search(ctx, index).Query(query).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SearchApi.Search``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchV1`: V1SearchResponse
	fmt.Fprintf(os.Stdout, "Response from `SearchApi.Search`: %+v\n", resp)
	for _, hit := range resp.Hits.Hits {
		fmt.Printf("%v %v\n", hit.GetTimestamp(), hit.GetSource())
	}
}

func TestClientInput(t *testing.T) {
	// string | Index
	id := uuid.NewString() // string | ID
	document := map[string]interface{}{
		"filename": "test.txt",
		"filepath": "/usr/local/test.txt",
		"content":  "你好 hello 再见 goodbye",
	} // map[string]interface{} | Document

	ctx := context.WithValue(context.Background(), zinc.ContextBasicAuth, zinc.BasicAuth{
		UserName: "admin",
		Password: "User#123",
	})
	configuration := zinc.NewConfiguration()
	configuration.Servers = zinc.ServerConfigurations{
		zinc.ServerConfiguration{
			URL: "http://localhost:4080",
		},
	}

	apiClient := zinc.NewAPIClient(configuration)
	resp, r, err := apiClient.Document.IndexWithID(ctx, index, id).Document(document).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `Document.IndexWithID``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IndexWithID`: MetaHTTPResponseID
	fmt.Fprintf(os.Stdout, "Response from `Document.IndexWithID`: %v\n", resp.GetId())
}

func TestDelete(t *testing.T) {
	docId := "id_example123"
	res, err := RpcServer.zincDelete(docId, index)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(res))
}
func TestQuery(t *testing.T) {
	res, err := RpcServer.zincQuery(QueryReq{
		SearchType: "querystring",
		Query: Query{
			Term: "zinc_test",
		},
		From:      0,
		MaxResult: 10,
	}, index)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("%v", res))
}

func TestPost(t *testing.T) {
	client := &http.Client{Timeout: time.Second * time.Duration(10)}
	resp, err := client.PostForm("http://127.0.0.1:6317/api/query", url.Values{"query": []string{"apiClient"}})
	if resp == nil {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	payloads, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	// fmt.Println(string(payloads))
	resultJson := gjson.Get(string(payloads), "data").String()
	var result QueryResp
	err = json.Unmarshal([]byte(resultJson), &result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v", result)
}

func TestFormatFilename(t *testing.T) {
	fmt.Println(formatFilename("NihaoTest_something.txt"))
}
