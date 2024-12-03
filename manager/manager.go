package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/araminian/cube/task"
	"github.com/araminian/cube/worker"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[uuid.UUID]*task.Task
	EventDb       map[uuid.UUID]*task.TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
}

func NewManager(workers []string) *Manager {
	taskDB := make(map[uuid.UUID]*task.Task)
	eventDB := make(map[uuid.UUID]*task.TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)

	for _, w := range workers {
		workerTaskMap[w] = []uuid.UUID{}
	}

	return &Manager{
		Pending:       *queue.New(),
		Workers:       workers,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
		TaskDb:        taskDB,
		EventDb:       eventDB,
	}
}

func (m *Manager) SelectWorker() string {
	var newWorker int
	if m.LastWorker+1 < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker = newWorker
	} else {
		newWorker = 0
		m.LastWorker = 0
	}
	return m.Workers[newWorker]
}

func (m *Manager) UpdateTasks() {
	for _, worker := range m.Workers {
		log.Printf("Manager: Checking worker %s for tasks updates", worker)
		url := fmt.Sprintf("http://%s/tasks", worker)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Manager: Error getting tasks from worker %s: %v", worker, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Manager: Non-200 response from worker %s: %d", worker, resp.StatusCode)
			continue
		}

		d := json.NewDecoder(resp.Body)
		var tasks []*task.Task
		err = d.Decode(&tasks)
		if err != nil {
			log.Printf("Manager: Error decoding tasks from worker %s: %v", worker, err)
			continue
		}

		for _, t := range tasks {
			log.Printf("Manager: Attempting to update task %v", t.ID)

			_, ok := m.TaskDb[t.ID]
			if !ok {
				log.Printf("Manager: Task %v not found in task db", t.ID)
				continue
			}

			if m.TaskDb[t.ID].State != t.State {
				m.TaskDb[t.ID].State = t.State
			}

			m.TaskDb[t.ID].StartTime = t.StartTime
			m.TaskDb[t.ID].FinishTime = t.FinishTime
			m.TaskDb[t.ID].ContainerID = t.ContainerID
		}
	}
}

func (m *Manager) SendWork() {
	if m.Pending.Len() > 0 {

		w := m.SelectWorker()

		e := m.Pending.Dequeue()
		te := e.(task.TaskEvent)
		t := te.Task
		log.Printf("Pulled %v off pending queue", t)

		m.EventDb[te.ID] = &te
		m.WorkerTaskMap[w] = append(m.WorkerTaskMap[w], t.ID)
		m.TaskWorkerMap[t.ID] = w

		t.State = task.Scheduled

		data, err := json.Marshal(te)
		if err != nil {
			log.Printf("Error marshalling task: %v", err)
		}

		url := fmt.Sprintf("http://%s/tasks", w)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Error sending work %v to worker %s: %v", te, w, err)
			m.Pending.Enqueue(te)
			return
		}
		defer resp.Body.Close()

		d := json.NewDecoder(resp.Body)

		if resp.StatusCode != http.StatusCreated {
			e := worker.ErrorResponse{}
			err := d.Decode(&e)
			if err != nil {
				log.Printf("Error decoding error response from worker %s: %v", w, err)
				return
			} else {
				log.Printf("Response %d Error from worker %s: %s", e.HTTPStatus, w, e.Message)
				return
				// Should i readd the task to pending queue?
			}
		}
		t = task.Task{}
		err = d.Decode(&t)
		if err != nil {
			log.Printf("Error decoding task response from worker %s: %v", w, err)
			return
		}
		log.Printf("Received task %v from worker %s", t, w)
		m.TaskDb[t.ID] = &t
	} else {
		log.Println("Manager: No tasks to send")
	}

}

func (m *Manager) AddTask(te task.TaskEvent) {
	m.Pending.Enqueue(te)
}
