package gopool

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
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

func TestPoolManyTrigged(t *testing.T) {
	var count int32 = 0
	job0 := NewJob("job0", &testJob{})
	job1 := NewJob("job1", &testJob{})
	job2 := NewJob("job2", &testJob{})
	job3 := NewJob("job3", &testJob{}).WithResultCallback(func(result interface{}, err error) {
		atomic.AddInt32(&count, 1)
	})

	p := NewPool(10, 3).WithPanicCallback(func(r interface{}) {
		t.Logf("%v", r)
	})

	if err := job0.Before(job1, job2); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	if err := job3.When(func(self *Job) bool {
		for _, job := range self.GetUpstreams() {
			if job.GetStatus() != JobSuccess {
				return false
			}
		}
		return true
	}).After(job1, job2); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	pipeline, err := NewPipeline("pipeline", job0, job1, job2, job3)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	if err := p.AddPipeline(pipeline); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	time.Sleep(2 * time.Second)
	job3TriggedCount := atomic.LoadInt32(&count)
	if job3TriggedCount != 1 && job3TriggedCount != 2 {
		t.Error("not execute")
		t.FailNow()
	}
	t.Logf("job3 trigged count %d", job3TriggedCount)
	p.Close("finish")
}
