package main

import (
	"fmt"
	"time"

	"github.com/araminian/cube/manager"
	"github.com/araminian/cube/node"
	"github.com/araminian/cube/task"
	"github.com/araminian/cube/worker"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {

	t := task.Task{
		ID:     uuid.New(),
		Name:   "test",
		State:  task.Pending,
		Image:  "nginx:latest",
		Memory: 1024,
		Disk:   1,
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		Timestamp: time.Now(),
		Task:      t,
	}

	fmt.Printf("task :=> %+v\n", t)
	fmt.Printf("task event :=> %+v\n", te)

	worker1 := worker.Worker{
		Name:  "worker1",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	fmt.Printf("worker1 :=> %+v\n", worker1)

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string][]*task.Task),
		EventDb: make(map[string][]*task.TaskEvent),
		Workers: []string{worker1.Name},
	}

	fmt.Printf("manager :=> %+v\n", m)

	n := node.Node{
		Name:   "node-1",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 8192,
		Disk:   100,
		Role:   "worker",
	}

	fmt.Printf("node :=> %+v\n", n)
}
