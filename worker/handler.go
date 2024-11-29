package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/araminian/cube/task"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ErrorResponse struct {
	Message    string `json:"message"`
	HTTPStatus int    `json:"http_status"`
}

func (a *API) StartTaskHandler(w http.ResponseWriter, r *http.Request) {

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var taskEvent task.TaskEvent
	err := d.Decode(&taskEvent)
	if err != nil {
		msg := fmt.Sprintf("Error decoding task: %s", err)
		log.Println(msg)
		e := ErrorResponse{
			Message:    msg,
			HTTPStatus: http.StatusBadRequest,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	a.Worker.AddTask(taskEvent.Task)
	log.Printf("Task %s added to worker %s\n", taskEvent.Task.ID, a.Worker.Name)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(taskEvent.Task)
}

func (a *API) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(a.Worker.GetTasks())

}

func (a *API) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		log.Println("No task ID provided")
		e := ErrorResponse{
			Message:    "No task ID provided",
			HTTPStatus: http.StatusBadRequest,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	tID, err := uuid.Parse(taskID)
	if err != nil {
		log.Printf("Invalid task ID: %s", taskID)
		e := ErrorResponse{
			Message:    "Invalid task ID",
			HTTPStatus: http.StatusBadRequest,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	_, ok := a.Worker.Db[tID]
	if !ok {
		log.Printf("Task %s not found", tID)
		e := ErrorResponse{
			Message:    "Task not found",
			HTTPStatus: http.StatusNotFound,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	taskToStop := a.Worker.Db[tID]

	taskCopy := *taskToStop

	taskCopy.State = task.Completed

	a.Worker.AddTask(taskCopy)

	log.Printf("Task %v with container %s stopped on worker %s\n", taskToStop, taskToStop.ContainerID, a.Worker.Name)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskToStop)

}
