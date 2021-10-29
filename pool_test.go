package gopool

import (
	"fmt"
	"testing"
)

const ()

func BenchmarkPoolJob(b *testing.B) {
	p := NewPool(200000, 1000)
	defer func() {
		p.Close("finish")
		b.StopTimer()
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		jobName := fmt.Sprintf("testjob-%d", i)
		for j := 0; j < 1000000; j++ {
			p.AddJob(NewJob(jobName, &testJob{Name: jobName}))
		}
	}
}

func TestPoolPanic(t *testing.T) {
	p := NewPool(10, 2).WithPanicCallback(func(r interface{}) {
		t.Logf("%v", r)
	}).WithEventCallback(EventLevelDebug, func(event *Event) {
		t.Log(event)
	})

	panicJob := NewJob("panicJob", &panicJob{})
	p.AddJob(panicJob, NewJob("normal job", &testJob{Name: "testJob"}))
	p.Close("finish")

}

func TestPoolResultCallback(t *testing.T) {
	p := NewPool(10, 2).WithPanicCallback(func(r interface{}) {
		t.Logf("%v", r)
	}).WithEventCallback(EventLevelDebug, func(event *Event) {
		t.Log(event)
	})
	p.AddJob(NewJob("resultCallback", &testJob{Name: "resultCallback"}).WithResultCallback(func(result interface{}, err error) {
		t.Logf("%v %v", result, err)
	}))

	p.Close("finish")
}
