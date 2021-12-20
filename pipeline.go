package goplumber

import (
	"context"
	"sync"

	"cloud.google.com/go/civil"
	"github.com/heimdalr/dag"
)

type Pipeline struct {
	dag       *dag.DAG
	Name      string
	tasks     map[string]*Task
	vertexIDs map[string]string
	Store     Store
}

type PipelineConfig struct {
	Store Store
}

// DefaultConfig creates a pipeline config. It use the default store.
func DefaultConfig() *PipelineConfig {
	return &PipelineConfig{
		Store: DefaultStore(),
	}
}

// NewPipeline creates a new pipeline.
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		dag:       dag.NewDAG(),
		Name:      name,
		tasks:     make(map[string]*Task),
		vertexIDs: make(map[string]string),
	}
}

// AddTasks adds multiple tasks to the pipeline.
func (p *Pipeline) AddTasks(tasks ...*Task) *Pipeline {
	for _, t := range tasks {
		v, _ := p.dag.AddVertex(t)
		p.vertexIDs[t.Name] = v
		p.tasks[v] = t
	}
	return p
}

// AddTask adds a task to the pipeline.
func (p *Pipeline) AddTask(t *Task) *Pipeline {
	return p.AddTasks(t)
}

// Depend sets the dependency between the two tasks.
func (p *Pipeline) Depend(prior *Task, subsequent *Task) *Pipeline {
	_ = p.dag.AddEdge(p.vertexIDs[prior.Name], p.vertexIDs[subsequent.Name])
	return p
}

func (p *Pipeline) processRunQueue(
	ctx context.Context,
	runq chan string,
	compq chan string,
	wg *sync.WaitGroup,
) {
	// Receive a task invoke message.
	for id := range runq {
		// Cannot start the task because there are still uncompleted prior tasks.
		if !p.isReady(id) {
			continue
		}
		// Start a task worker as a goroutine.
		go func(ctx context.Context, id string) {
			wg.Add(1)
			t := p.tasks[id]
			if !t.IsCompleted() {
				t.transition(ctx, Running)
				var params interface{}
				if _, err := t.Do(ctx, params); err != nil {
					// TODO: how to handle the error
					// should i hundle the multiple errors?
					panic(err)
				}
				t.transition(ctx, Completed)
			}
			// Send a message of completed task.
			compq <- id
			wg.Done()
		}(ctx, id)
	}
	wg.Done()
}

func (p *Pipeline) processCompletionQueue(
	ctx context.Context,
	runq chan string,
	compq chan string,
	wg *sync.WaitGroup,
) {
	for id := range compq {
		t := p.tasks[id]
		if t.IsCompleted() {
			// Terminate this pipeline when all tasks are completed.
			if p.CountUncompletedTasks() == 0 {
				break
			}

			// Invoke subsequent tasks.
			r, _ := p.dag.GetChildren(id)
			for cid, _ := range r {
				runq <- cid
			}
		}
	}

	// Close channels.
	close(runq)
	close(compq)

	wg.Done()
}

func (p *Pipeline) CountUncompletedTasks() int64 {
	c := int64(0)
	for _, t := range p.tasks {
		if !t.IsCompleted() {
			c++
		}
	}
	return c
}

type Schedule string

const (
	None   Schedule = "None"
	Once   Schedule = "Once"
	Daily  Schedule = "Daily"
	Hourly Schedule = "Hourly"
)

type KeyTransformer interface {
	Transform(id string) string
}

type DailyKeyTransformer struct {
	DS civil.Date
	KeyTransformer
}

/*
func (t *DailyKeyTransformer) Transform(id string) string {
	return fmt.Sprintf("%s/%s", t.DS.String(), id)
}

func (p *Pipeline) Book(ctx context.Context) error {
	for _, task := range p.tasks {
		key := p.KeyTransformer.Transform(task.ID())

		t := &Task{}
		found, err := p.Store.Get(key, t)
		if err != nil {
			return err
		}

		if found {
			continue
		}

		if err := p.Store.Set(key, task); err != nil {
			return err
		}
	}

	return nil
}
*/

func (p *Pipeline) Run(ctx context.Context) error {
	runq := make(chan string)
	compq := make(chan string)
	wg := &sync.WaitGroup{}

	for id, _ := range p.dag.GetRoots() {
		runq <- id
	}

	wg.Add(1)
	go p.processRunQueue(ctx, runq, compq, wg)

	wg.Add(1)
	go p.processCompletionQueue(ctx, runq, compq, wg)

	wg.Wait()

	return nil
}

func (p *Pipeline) isReady(id string) bool {
	ancestors, _ := p.dag.GetAncestors(id)
	for _, v := range ancestors {
		t := v.(Task)
		if t.Status != Completed {
			return false
		}
	}
	return true
}
