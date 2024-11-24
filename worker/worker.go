package worker

import (
	"github.com/araminian/cube/task"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
}

func (w *Worker) CollectStats() {}

func (w *Worker) RunTask() {}

func (w *Worker) StartTask() {}

func (w *Worker) StopTask() {}
