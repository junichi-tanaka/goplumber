package goplumber

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/heimdalr/dag"
)

func TestNewTask(t *testing.T) {
	h := func(context.Context, interface{}) (interface{}, error) {
		return nil, nil
	}

	type args struct {
		name    string
		handler Handler
	}
	tests := []struct {
		name string
		args args
		want *Task
	}{
		{
			args: args{
				name:    "Test Task",
				handler: h,
			},
			want: &Task{
				Name:    "Test Task",
				handler: h,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := NewTask(tt.args.name, tt.args.handler)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(Task{})); diff != "" {
				t.Errorf("NewTask() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTask_ID(t *testing.T) {
	type fields struct {
		IDInterface dag.IDInterface
		Name        string
		Status      TaskStatus
		handler     Handler
		m           *sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{
				Name: "Test Task",
			},
			want: "Test Task",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				IDInterface: tt.fields.IDInterface,
				Name:        tt.fields.Name,
				Status:      tt.fields.Status,
				handler:     tt.fields.handler,
				m:           tt.fields.m,
			}
			if got := tr.ID(); got != tt.want {
				t.Errorf("Task.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_Do(t *testing.T) {
	type fields struct {
		IDInterface dag.IDInterface
		Name        string
		Status      TaskStatus
		handler     Handler
		m           *sync.Mutex
	}
	type args struct {
		ctx    context.Context
		params interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "add",
			fields: fields{
				handler: func(ctx context.Context, params interface{}) (interface{}, error) {
					v := params.(int)
					return v + 1, nil
				},
				m: &sync.Mutex{},
			},
			args: args{
				ctx:    context.Background(),
				params: 1,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				handler: func(ctx context.Context, params interface{}) (interface{}, error) {
					return nil, errors.New("error")
				},
				m: &sync.Mutex{},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				IDInterface: tt.fields.IDInterface,
				Name:        tt.fields.Name,
				Status:      tt.fields.Status,
				handler:     tt.fields.handler,
				m:           tt.fields.m,
			}
			got, err := tr.Do(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Task.Do() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_transition(t *testing.T) {
	type fields struct {
		IDInterface dag.IDInterface
		Name        string
		Status      TaskStatus
		handler     Handler
		m           *sync.Mutex
	}
	type args struct {
		ctx    context.Context
		status TaskStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				IDInterface: tt.fields.IDInterface,
				Name:        tt.fields.Name,
				Status:      tt.fields.Status,
				handler:     tt.fields.handler,
				m:           tt.fields.m,
			}
			if err := tr.transition(tt.args.ctx, tt.args.status); (err != nil) != tt.wantErr {
				t.Errorf("Task.transition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTask_IsCompleted(t *testing.T) {
	type fields struct {
		IDInterface dag.IDInterface
		Name        string
		Status      TaskStatus
		handler     Handler
		m           *sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			fields: fields{
				Status: Completed,
				m:      &sync.Mutex{},
			},
			want: true,
		},
		{
			fields: fields{
				Status: Running,
				m:      &sync.Mutex{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				IDInterface: tt.fields.IDInterface,
				Name:        tt.fields.Name,
				Status:      tt.fields.Status,
				handler:     tt.fields.handler,
				m:           tt.fields.m,
			}
			if got := tr.IsCompleted(); got != tt.want {
				t.Errorf("Task.IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}
