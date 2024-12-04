package main

import (
	"fmt"
	"log"
	"time"

	"github.com/araminian/cube/manager"
	"github.com/araminian/cube/task"
	"github.com/araminian/cube/worker"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {

	db := make(map[uuid.UUID]*task.Task)
	w := worker.Worker{
		Name:  "worker1",
		Queue: *queue.New(),
		Db:    db,
	}

	whost := "localhost"
	wport := 5555

	mhost := "localhost"
	mport := 5556

	wapi := worker.API{
		Worker:  &w,
		Address: whost,
		Port:    wport,
	}

	go w.RunTask()
	go w.CollectStats()
	go wapi.Start()

	workers := []string{fmt.Sprintf("%s:%d", whost, wport)}
	fmt.Printf("Manager : Starting with workers %v\n", workers)
	m := manager.NewManager(workers)
	mapi := manager.Api{
		Manager: m,
		Address: mhost,
		Port:    mport,
	}

	go m.ProcessTasks()
	go m.UpdateTasks()
	go mapi.Start()

	for {
		log.Printf("Manager: TaskDB: %v\n", m.TaskDb)
		for _, t := range m.TaskDb {
			fmt.Printf("Manager Task : id: %s , state: %d\n", t.ID, t.State)
		}
		time.Sleep(10 * time.Second)
	}
}

// Worker
// curl -X POST http://localhost:5555/tasks -d '{"ID":"123e4567-e89b-12d3-a456-426614174000","State":2,"TASK":{"ID":"123e4567-e89b-12d3-a456-426614174000","State":1,"Name":"test","Image":"nginx:latest"}}'
// curl localhost:5555/tasks
// curl -X DELETE localhost:5555/tasks/123e4567-e89b-12d3-a456-426614174000

// Manager
// curl -X POST http://localhost:5556/tasks -d '{"ID":"123e4567-e89b-12d3-a456-426614174000","State":2,"TASK":{"ID":"123e4567-e89b-12d3-a456-426614174000","State":1,"Name":"test","Image":"nginx:latest"}}'
// curl localhost:5556/tasks
// curl -X DELETE localhost:5556/tasks/123e4567-e89b-12d3-a456-426614174000
