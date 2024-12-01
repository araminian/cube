package worker

import (
	"fmt"
	"log"
	"time"

	"github.com/araminian/cube/task"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
	Stats     Stats
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) RunTask() task.DockerResult {
	fromQ := w.Queue.Dequeue()

	if fromQ == nil {
		log.Printf("No task to in the queue")
		return task.DockerResult{Error: nil}
	}

	taskQueued := fromQ.(task.Task)

	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = taskPersisted
	}

	var result task.DockerResult

	if task.ValidateStateTransition(
		taskPersisted.State,
		taskQueued.State,
	) {
		switch taskQueued.State {
		case task.Scheduled:
			result = w.StartTask(taskQueued)
		case task.Completed:
			result = w.StopTask(taskQueued)
		default:
			log.Printf("Invalid state transition for task %s", taskQueued.ID)
		}
	} else {
		err := fmt.Errorf("invalid state transition from %v to %v for task %s", taskPersisted.State, taskQueued.State, taskQueued.ID)
		result.Error = err
	}

	return result
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {

	t.StartTime = time.Now()

	config := task.NewConfig(&t)

	docker, err := task.NewDocker(config)
	if err != nil {
		log.Printf("Error creating docker: %+v", err)
		return task.DockerResult{Error: err}
	}

	result := docker.Run()
	if result.Error != nil {
		log.Printf("Error running task %s: %+v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerID
	t.State = task.Running
	w.Db[t.ID] = &t

	return result
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {

	config := task.NewConfig(&t)

	docker, err := task.NewDocker(config)
	if err != nil {
		log.Printf("Error creating docker: %+v", err)
		return task.DockerResult{Error: err}
	}

	result := docker.Stop(t.ContainerID)

	if result.Error != nil {
		log.Printf("Error stopping docker with id %s: %+v", t.ContainerID, result.Error)
		return result
	}

	t.FinishTime = time.Now()
	t.State = task.Completed
	w.Db[t.ID] = &t

	log.Printf("Stopped and Removed container %s for task %s", t.ContainerID, t.ID)

	return result
}

func (w *Worker) GetTasks() []task.Task {
	tasks := []task.Task{}
	for _, t := range w.Db {
		tasks = append(tasks, *t)
	}
	return tasks
}

func (w *Worker) CollectStats() {
	for {
		log.Printf("Collecting stats in %s", w.Name)
		w.Stats = getStats()
		w.Stats.TaskCount = w.TaskCount
		time.Sleep(15 * time.Second)
	}
}
