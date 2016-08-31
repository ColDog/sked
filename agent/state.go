package agent

import (
	"github.com/coldog/sked/api"
	"sync"
	"time"
	"encoding/json"
	"io/ioutil"
)

type TaskState struct {
	StartedAt time.Time `json:"started_at"`
	Attempts  int       `json:"attempts"`
	Failure   error     `json:"failure"`
	Healthy   bool      `json:"healthy"`
	Task      *api.Task `json:"task"`
}

type AgentState struct {
	State map[string]*TaskState `json:"state"`
	l     *sync.RWMutex
}

func (a *AgentState) get(key string, tasks ...*api.Task) *TaskState {
	a.l.RLock()
	t, ok := a.State[key]
	a.l.RUnlock()

	if !ok {
		t = &TaskState{}
		a.State[key] = t
	}

	if len(tasks) == 1 && t.Task == nil {
		t.Task = tasks[0]
	}

	return t
}

func (a *AgentState) del(key string) {
	a.l.RLock()
	_, ok := a.State[key]
	a.l.RUnlock()

	if ok {
		a.l.Lock()
		delete(a.State, key)
		a.l.Unlock()
	}
}

func (a *AgentState) save() error {
	res, err := json.Marshal(a.State)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("agent-state.json", res, 0644)
}

func (a *AgentState) load() error {
	data, err := ioutil.ReadFile("agent-state.json")
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &a.State)
}