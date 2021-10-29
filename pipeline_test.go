package gopool

import "testing"

func TestPipelineGraph(t *testing.T) {
	job1 := NewJob("job1", &testJob{Name: "job1"})
	job2 := NewJob("job2", &testJob{Name: "job2"})
	job3 := NewJob("job3", &testJob{Name: "job3"})
	job4 := NewJob("job4", &testJob{Name: "job4"})
	if err := job2.After(job1); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	job3.After(job2)
	job4.Before(job3)
	pipeline, err := NewPipeline("job-pipeline", job1, job2, job3, job4)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	graph, err := pipeline.Graph()
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	t.Log(graph)
}
