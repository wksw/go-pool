package gopool

import "testing"

type testJob struct {
	Name string
}

type testJob1 struct{}

func (j *testJob) Handle() (interface{}, error) {
	return nil, nil
}

func (j *testJob1) Handle() (interface{}, error) {
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

func TestAddBefore(t *testing.T) {
	job1 := NewJob("test_job1", &testJob{Name: "test1"})
	job2 := NewJob("test_job2", &testJob{Name: "test2"})
	if err := job2.Before(job1); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestAddAfter(t *testing.T) {
	job1 := NewJob("test_job1", &testJob{Name: "test1"})
	job2 := NewJob("test_job2", &testJob{Name: "test2"})
	if err := job2.After(job1); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestAddCycle(t *testing.T) {
	job1 := NewJob("test_job1", &testJob{Name: "test1"})
	job2 := NewJob("test_job2", &testJob{Name: "test2"})
	job3 := NewJob("test_job3", &testJob{Name: "test3"})
	if err := job1.Before(job2); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if err := job3.After(job2); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if err := job1.After(job3); err == nil {
		t.FailNow()
	}
	_, err := NewPipeline("job_test", job1, job2, job3)
	t.Log(err)
	if err == nil {
		t.FailNow()
	}

}
