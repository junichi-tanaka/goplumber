package goplumber_test

import (
	"context"
	"fmt"
	"time"

	"github.com/junichi-tanaka/goplumber"
	gokvfile "github.com/philippgille/gokv/file"
)

func ExampleTest1() {
	store, _ := gokvfile.NewStore(gokvfile.DefaultOptions)
	defer store.Close()

	p := goplumber.NewPipeline("test")

	t1 := goplumber.NewTask("task1", func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("task1")
		time.Sleep(1 * time.Second)
		return nil, nil
	})
	t2 := goplumber.NewTask("task2", func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("task2")
		time.Sleep(1 * time.Second)
		return nil, nil
	})
	t3 := goplumber.NewTask("task3", func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("task3")
		time.Sleep(1 * time.Second)
		return nil, nil
	})
	t4 := goplumber.NewTask("task4", func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("task4")
		time.Sleep(2 * time.Second)
		return nil, nil
	})

	p.AddTask(t1)
	p.AddTask(t2)
	p.AddTask(t3)
	p.AddTask(t4)

	p.Depend(t1, t2)
	p.Depend(t2, t3)
	p.Depend(t2, t4)

	p.Run(context.Background())
}
