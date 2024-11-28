package main

import (
	"fmt"
	"log"
	"time"

	"github.com/araminian/cube/task"
	"github.com/araminian/cube/worker"
	"github.com/docker/docker/client"
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

	tt := task.Task{
		ID:    uuid.New(),
		Name:  "test",
		State: task.Scheduled,
		Image: "nginx:latest",
	}

	fmt.Printf("task :=> %+v\n", tt)
	w.AddTask(tt)

	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

	tt.ContainerID = result.ContainerID
	fmt.Printf("task %s with container id %s\n", tt.ID, tt.ContainerID)

	fmt.Printf("Sleeping for 1min\n")
	time.Sleep(1 * time.Minute)

	tt.State = task.Completed
	w.AddTask(tt)
	result = w.RunTask()
	if result.Error != nil {
		panic(result.Error)
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
