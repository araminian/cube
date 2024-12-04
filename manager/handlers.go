package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/araminian/cube/task"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ErrResponse struct {
	HTTPStatusCode int    `json:"status"`
	Message        string `json:"message"`
}

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	te := task.TaskEvent{}
	err := d.Decode(&te)
	if err != nil {
		msg := fmt.Sprintf("failed to decode task event: %v", err)
		log.Println(msg)
		e := ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	a.Manager.AddTask(te)
	log.Printf("Manager: task added: %+v", te)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(te.Task)
}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	tasks := a.Manager.GetTasks()
	json.NewEncoder(w).Encode(tasks)
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {

	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		msg := "taskID is required"
		log.Println(msg)
		e := ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	tID, _ := uuid.Parse(taskID)

	taskToStop, ok := a.Manager.TaskDb[tID]
	if !ok {
		msg := fmt.Sprintf("task not found: %s", taskID)
		log.Println(msg)
		e := ErrResponse{
			HTTPStatusCode: http.StatusNotFound,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Completed,
		Timestamp: time.Now(),
	}

	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	te.Task = taskCopy

	a.Manager.AddTask(te)

	log.Printf("Manager: Added task event to stop task %s: %+v", taskID, te)
	w.WriteHeader(http.StatusNoContent)
}
