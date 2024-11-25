package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/araminian/cube/manager"
	"github.com/araminian/cube/node"
	"github.com/araminian/cube/task"
	"github.com/araminian/cube/worker"
	"github.com/docker/docker/client"
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

	fmt.Println("Creating container")
	dockerTask, createResult := createContainer()

	if createResult.Error != nil {
		fmt.Printf("Error creating container: %v", createResult.Error)
		os.Exit(1)
	}

	time.Sleep(10 * time.Second)

	fmt.Printf("Stopping container %s with id %s", dockerTask.Config.Name, createResult.ContainerID)

	stopResult := stopContainer(dockerTask, createResult.ContainerID)

	if stopResult.Error != nil {
		fmt.Printf("Error stopping container: %v", stopResult.Error)
		os.Exit(1)
	}
}

func createContainer() (*task.Docker, *task.DockerResult) {

	c := task.Config{
		Image: "nginx:latest",
		Name:  "test",
		Env:   []string{"ENV=test"},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)

	d := task.Docker{
		Client: dc,
		Config: c,
	}

	result := d.Run()

	if result.Error != nil {
		log.Printf("Error creating container: %v", result.Error)
		return nil, nil
	}

	fmt.Printf("Container created: %s", result.ContainerID)

	return &d, &result
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	result := d.Stop(id)

	if result.Error != nil {
		log.Printf("Error stopping container: %v", result.Error)
		return nil
	}

	fmt.Printf("Container stopped and removed: %s", id)

	return &result
}
