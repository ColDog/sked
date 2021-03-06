package api

import (
	"github.com/coldog/sked/tools"
	"testing"
)

func TestExecutors_Bash(t *testing.T) {
	b := BashExecutor{
		Start: "echo 'hello'",
		Stop:  "echo 'stop'",
	}

	task := SampleTask()
	err := b.StartTask(task)
	tools.Ok(t, err)

	err = b.StopTask(task)
	tools.Ok(t, err)
}

func TestExecutors_Docker(t *testing.T) {
	b := DockerExecutor{
		Image:         "ubuntu",
		ContainerPort: 8080,
		Env:           []string{"HI=hello"},
	}

	task := SampleTask()
	err := b.StartTask(task)
	tools.Ok(t, err)

	err = b.StopTask(task)
	tools.Ok(t, err)
}
