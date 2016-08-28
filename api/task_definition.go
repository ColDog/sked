package api

import (
	"encoding/json"
	"github.com/coldog/sked/tools"
	"time"
)

// a runnable task description
// validations:
// 	- version must be new
//
type TaskDefinition struct {
	Name        string        `json:"name"`
	Version     uint          `json:"version"`
	ProvidePort bool          `json:"provide_port"`
	Port        uint          `json:"port"`
	Tags        []string      `json:"tags"`
	Containers  []*Container  `json:"containers"`
	GracePeriod time.Duration `json:"grace_period"`
}

func (task *TaskDefinition) Validate(api SchedulerApi) (errors []string) {
	_, err := api.GetTaskDefinition(task.Name, task.Version)
	if err == nil {
		errors = append(errors, "version already provisioned")
	}
	return errors
}

func (t *TaskDefinition) Key() string {
	return "config/task/" + t.Name
}

type Container struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Executor json.RawMessage `json:"executor"`
	Setup    []string        `json:"setup"`
	Checks   []*Check        `json:"checks"`
	Memory   uint64          `json:"memory"`
	CpuUnits uint64          `json:"cpu_units"`
	DiskUse  uint64          `json:"disk_use"`
	bash     *BashExecutor
	docker   *DockerExecutor
}

func (c *Container) RunSetup() error {
	for _, cmd := range c.Setup {
		err := tools.Exec(nil, "/bin/bash", "-c", cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

// a check passed along to consul
type Check struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	HTTP     string `json:"http"`
	TCP      string `json:"tcp"`
	Script   string `json:"script"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
	TTL      string `json:"ttl"`
}
