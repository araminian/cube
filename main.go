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

	host := "localhost"
	port := 5555

	api := worker.API{
		Worker:  &w,
		Address: host,
		Port:    port,
	}

	go runTasks(&w)
	go w.CollectStats()
	api.Start()
}

func runTasks(w *worker.Worker) {
	for {
		if w.Queue.Len() > 0 {
			result := w.RunTask()
			if result.Error != nil {
				log.Printf("Error running task: %v", result.Error)
			}
		} else {
			log.Printf("No tasks to run on worker %s\n", w.Name)
		}
		log.Printf("Sleeping for 10 seconds\n")
		time.Sleep(10 * time.Second)
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

// curl -X POST http://localhost:5555/tasks -d '{"ID":"123e4567-e89b-12d3-a456-426614174000","State":2,"TASK":{"ID":"123e4567-e89b-12d3-a456-426614174000","State":1,"Name":"test","Image":"nginx:latest"}}'
// curl localhost:5555/tasks
// curl -X DELETE localhost:5555/tasks/123e4567-e89b-12d3-a456-426614174000
