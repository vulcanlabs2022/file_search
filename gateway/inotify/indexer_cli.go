package inotify

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"wzinc/common"
	"wzinc/rpc"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	DeleteAction = "delete"
	AddAction    = "add"
	UpdataAction = "update"
)

var VectorCli BaseClient

var IndexerUrl string

func init() {
	VectorCli = BaseClient{
		taskCallback: make(chan VectorDBTaskStatus),
		fsTask:       make(chan VectorDBTask),
		taskList: taskList{
			eleMap: map[string]*taskElement{},
			root: &taskElement{
				task: nil,
				next: nil,
				prev: nil,
			},
		},
		pendingTask: make(map[string]*VectorDBTask),
	}
}

type BaseClient struct {
	taskCallback chan VectorDBTaskStatus
	fsTask       chan VectorDBTask
	taskList     taskList
	pendingTask  map[string]*VectorDBTask //taskId -> task
}

func callIndexerTest(task *VectorDBTask) ([]byte, error) {
	log.Info().Msgf("call %s task taskId %v name %v path %s", task.Action, task.TaskId, task.Filename, task.Filepath)
	return []byte("ok"), nil
}

func callIndexer(task *VectorDBTask) ([]byte, error) {
	log.Info().Msgf("call %s task taskId %v name %v path %s", task.Action, task.TaskId, task.Filename, task.Filepath)
	b, _ := json.Marshal(task)
	log.Debug().Msgf("call indexer body %s", string(b))
	resp, err := common.HttpPost(IndexerUrl, string(b), 60)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func (bc *BaseClient) Run() {
	bc.SubscribeTaskStatus()

	go func() {
		for {
			task := <-bc.fsTask
			bc.taskList.Push(task)
		}
	}()

	pollingDuration := time.Second * 5
	maxTaskDuration := time.Minute * 30
	timer := time.NewTimer(pollingDuration)

	callTask := func() {
		if nextTask, ok := bc.taskList.Pop(); ok {
			bc.pendingTask[nextTask.TaskId] = nextTask
			b, err := callIndexer(nextTask)
			if err != nil {
				log.Error().Msgf("call indexer task %v error %v", nextTask, err)
				bc.taskList.Push(*nextTask)
				timer.Reset(pollingDuration)
			} else {
				log.Debug().Msgf("call indexer task response %s", string(b))
				timer.Reset(maxTaskDuration)
			}
			return
		}
		// log.Debug().Msg("no task wait for next round")
		timer.Reset(pollingDuration)
	}

	for {
		select {
		case taskStatus := <-bc.taskCallback:
			log.Info().Msgf("task status %v", taskStatus)
			if taskStatus.StatusCode == TaskDone {
				log.Info().Msgf("task %s done", taskStatus.TaskId)
				delete(bc.pendingTask, string(taskStatus.TaskId))
			} else {
				log.Warn().Msgf("task %s failed", taskStatus.TaskId)
				if failedTask, ok := bc.pendingTask[string(taskStatus.TaskId)]; ok {
					log.Info().Msgf("push failed task to retry %v", *failedTask)
					bc.taskList.Push(*failedTask)
				}
			}
			callTask()
		case <-timer.C:
			callTask()
		}
	}
}

func (bc *BaseClient) SubscribeTaskStatus() error {
	rpc.RpcServer.CallbackGroup.POST("/vector", func(c *gin.Context) {
		data, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		var taskStatus VectorDBTaskStatus
		err = json.Unmarshal(data, &taskStatus)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		log.Info().Msgf("vector task callback %v", taskStatus)
		go func() {
			bc.taskCallback <- taskStatus
		}()
		c.String(http.StatusOK, "ok")
	})
	return nil
}
