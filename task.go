package goplumber

import (
	"context"
	"sync"

	"github.com/heimdalr/dag"
)

type TaskStatus int64

const (
	Pending   TaskStatus = 0
	Running   TaskStatus = 1
	Completed TaskStatus = 9
)

type Handler func(context.Context, interface{}) (interface{}, error)

type Task struct {
	dag.IDInterface
	Name    string
	Status  TaskStatus
	Done    bool
	handler Handler
	m       *sync.Mutex
}

func NewTask(name string, handler Handler) *Task {
	return &Task{
		Name:    name,
		handler: handler,
		m:       &sync.Mutex{},
	}
}

func (t *Task) ID() string {
	return t.Name
}

func (t *Task) Do(ctx context.Context, params interface{}) (interface{}, error) {
	r, err := t.handler(ctx, params)
	return r, err
}

func (t *Task) transition(ctx context.Context, status TaskStatus) error {
	t.m.Lock()
	defer t.m.Unlock()
	t.Status = status
	return nil
}

func (t *Task) IsCompleted() bool {
	t.m.Lock()
	defer t.m.Unlock()
	return t.Status == Completed
}
