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
