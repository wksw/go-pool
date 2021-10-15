package main

import (
	"fmt"
	"time"

	gopool "github.com/wksw/go-pool"
)

type job struct {
	Name string
}

var _ gopool.JobHandler = &job{}

// Handle job handler
func (j *job) Handle() (interface{}, error) {
	fmt.Println("job", j.Name, "handle")
	time.Sleep(100 * time.Millisecond)
	// panic("---abc")
	return nil, nil
}

func main() {
	pool := gopool.NewPool(100, 4).
		WithExitCallback(func(reason string) {
			fmt.Println("pool exit because", reason)
		}).
		WithPanicCallback(func(r interface{}) {
			fmt.Println("panic", r)
		}).
		WithEventCallback(gopool.EventLevelDebug, func(event *gopool.Event) {
			fmt.Println(event)
		})

	for i := 0; i < 10; i++ {
		pool.AddJob(gopool.NewJob("job", &job{Name: fmt.Sprintf("job-%d", i)}))
		fmt.Println("job ", i, "added")
	}
	pool.Close("finish")

}
