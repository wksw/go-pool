package gopool

import "testing"

type testJob struct {
	Name string
}

func (j *testJob) Handle() (interface{}, error) {
	return nil, nil
}

func TestDumplicateAdd(t *testing.T) {
	job := NewJob("test_job", &testJob{Name: "test"})
	err := job.After(job)
	if err == nil {
		t.FailNow()
	}
}

func TestCycleAdd(t *testing.T) {
	job1 := NewJob("test_job1", &testJob{Name: "test1"})
	job2 := NewJob("test_job2", &testJob{Name: "test2"})
	job2.After(job1)
	if err := job1.After(job2); err == nil {
		t.FailNow()
	}
}
