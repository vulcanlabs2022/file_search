package inotify

import (
	"crypto/sha256"
	"encoding/base32"
)

type VectorDBTask struct {
	Filename  string `json:"filename"`
	Filepath  string `json:"filepath"`
	IsInsert  bool   `json:"is_insert"`
	Action    string `json:"action"`
	TaskId    string `json:"task_id"`
	StartTime int64  `json:"startTime"`
	FileId    string `json:"file_id"`
}

type VectorDBTaskStatus struct {
	StatusCode int            `json:"status"`
	TaskId     VectorDBTaskId `json:"task_id"`
}

type VectorDBTaskId string

const (
	TaskDone  = 0
	TaskFaild = -1
)

type taskElement struct {
	task *VectorDBTask
	next *taskElement
	prev *taskElement
}

type taskList struct {
	eleMap map[string]*taskElement
	root   *taskElement
}

func (l *taskList) Pop() (*VectorDBTask, bool) {
	if l.root.next == nil {
		return nil, false
	}
	task := l.root.next.task
	l.delete(l.root.next)
	return task, task != nil
}

func (l *taskList) Push(newTask VectorDBTask) {
	//same file operation exist pending
	if pendingTask, ok := l.eleMap[newTask.Filepath]; ok {
		//incase pushed failed task
		if pendingTask.task.StartTime > newTask.StartTime {
			return
		}
		if pendingTask.task.IsInsert {
			if newTask.IsInsert {
				pendingTask.task.StartTime = newTask.StartTime
				return
			} else {
				l.delete(pendingTask)
				return
			}
		}
		if !pendingTask.task.IsInsert {
			if newTask.IsInsert {
				pendingTask.task = &newTask
				return
			} else {
				pendingTask.task.StartTime = newTask.StartTime
				return
			}
		}
	}
	//no pending operations
	l.add(&newTask)
}

func (l *taskList) delete(e *taskElement) {
	delete(l.eleMap, e.task.Filepath)
	if e.prev == nil {
		l.root.next = e.next
	} else {
		e.prev.next = e.next
	}
	if e.next == nil {
		l.root.prev = e.prev
	} else {
		e.next.prev = e.prev
	}
	e.next = nil
	e.prev = nil
}

func (l *taskList) add(task *VectorDBTask) *taskElement {
	e := &taskElement{
		task: task,
		next: nil,
		prev: nil,
	}
	l.eleMap[task.Filepath] = e
	if l.root.prev == nil {
		l.root.next = e
		l.root.prev = e
		return e
	}
	e.prev = l.root.prev
	l.root.prev.next = e
	l.root.prev = e
	return e
}

func fileId(filePath string) string {
	hash := sha256.Sum256([]byte(filePath))
	return base32.StdEncoding.EncodeToString(hash[:])
}
