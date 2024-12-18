package task

import (
	"context"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Failed
	Completed
)

var StateTransitionMap = map[State][]State{
	Pending:   {Scheduled},
	Scheduled: {Scheduled, Running, Failed},
	Running:   {Running, Completed, Failed},
	Completed: {},
	Failed:    {},
}

type Task struct {
	ID            uuid.UUID
	Name          string
	ContainerID   string
	State         State
	Image         string
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

func NewConfig(t *Task) Config {
	return Config{
		Name:          t.Name,
		Image:         t.Image,
		Memory:        int64(t.Memory),
		Disk:          int64(t.Disk),
		RestartPolicy: t.RestartPolicy,
		ExposedPorts:  t.ExposedPorts,
	}
}

type Docker struct {
	Client *client.Client
	Config Config
}

func NewDocker(config Config) (Docker, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return Docker{}, err
	}
	return Docker{
		Client: dc,
		Config: config,
	}, nil
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerID string
	Result      string
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(
		ctx, d.Config.Image, image.PullOptions{})

	if err != nil {
		log.Printf("Error pulling image %s: %v", d.Config.Image, err)
		return DockerResult{
			Error: err,
		}
	}
	io.Copy(os.Stdout, reader)
	rp := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	r := container.Resources{
		Memory:   int64(d.Config.Memory),
		NanoCPUs: int64(d.Config.Cpu * math.Pow(10, 9)),
	}

	cc := container.Config{
		Image: d.Config.Image,
		Tty:   false,
		Env:   d.Config.Env,
		Cmd:   d.Config.Cmd,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(
		ctx,
		&cc,
		&hc,
		nil,
		nil,
		d.Config.Name,
	)

	if err != nil {
		log.Printf("Error creating container %s from image %s: %v", d.Config.Name, d.Config.Image, err)
		return DockerResult{
			Error: err,
		}
	}

	err = d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("Error starting container %s: %v", resp.ID, err)
		return DockerResult{
			Error: err,
		}
	}

	// d.config.Runtime.ContainerID = resp.ID

	out, err := d.Client.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	if err != nil {
		log.Printf("Error getting logs for container %s: %v", resp.ID, err)
		return DockerResult{
			Error: err,
		}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{
		ContainerID: resp.ID,
		Action:      "start",
		Result:      "success",
		Error:       nil,
	}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Stopping container %s", id)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		log.Printf("Error stopping container %s: %v", id, err)
		return DockerResult{
			Error: err,
		}
	}

	err = d.Client.ContainerRemove(ctx, id, container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})
	if err != nil {
		log.Printf("Error removing container %s: %v", id, err)
		return DockerResult{
			Error: err,
		}
	}

	return DockerResult{
		Action: "stop",
		Result: "success",
		Error:  nil,
	}
}

func Contains(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func ValidateStateTransition(from State, to State) bool {
	validStates, ok := StateTransitionMap[from]
	if !ok {
		return false
	}
	return Contains(validStates, to)
}
