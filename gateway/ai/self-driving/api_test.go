package selfdriving

import (
	"fmt"
	"testing"
	"wzinc/common"
)

func TestGetAnswerFile(t *testing.T) {
	resp, err := common.HttpPostFile("", MaxPostTimeOut, map[string]string{
		common.PostFileParamKey:  "",
		common.PostQueryParamKey: "hello",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))
}
