package inotify

import (
	"encoding/json"
	"fmt"
	"path"
	"testing"
	"time"
)

func TestTaskList(t *testing.T) {

	VectorCli.taskList.Push(VectorDBTask{
		Filename:  "11",
		Filepath:  "11",
		IsInsert:  true,
		StartTime: 0,
	})

	VectorCli.taskList.Push(VectorDBTask{
		Filename:  "22",
		Filepath:  "22",
		IsInsert:  true,
		StartTime: 0,
	})

	VectorCli.taskList.Push(VectorDBTask{
		Filename:  "22",
		Filepath:  "22",
		IsInsert:  false,
		StartTime: 0,
	})

	task, ok := VectorCli.taskList.Pop()
	fmt.Println(ok)
	if ok {
		fmt.Printf("%v\n", *task)
	}

	task, ok = VectorCli.taskList.Pop()
	fmt.Println(ok)
	if ok {
		fmt.Printf("%v\n", *task)
	}

}

func TestTaskJson(t *testing.T) {
	filePath := "/Users/houmingyu/Documents/web5/filesearch/data/test/aaa.txt"
	task := VectorDBTask{
		Filename:  path.Base(filePath),
		Filepath:  filePath,
		IsInsert:  false,
		Action:    "add",
		TaskId:    "deecc6a4-bc2f-4c7d-8936-8659bfc24d88",
		StartTime: time.Now().Unix(),
		FileId:    fileId(filePath),
	}
	b, _ := json.Marshal(&task)
	fmt.Println(string(b))
}
