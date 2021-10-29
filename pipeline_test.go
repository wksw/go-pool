package gopool

import (
	"testing"
	"time"
)

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

func next(self *Job) bool {
	for _, job := range self.GetUpstreams() {
		if job.GetStatus() != JobSuccess {
			return false
		}
	}
	return true
}

func TestPipeline(t *testing.T) {
	pool := NewPool(10, 5).WithEventCallback(EventLevelDebug, func(event *Event) {
		t.Log(event)
	}).WithExitCallback(func(reason string) {
		t.Log(reason)
	})
	nameSpaceJob := NewJob("nameSpaceJob", &testJob{Name: "nameSpaceJob"})
	configmapJob := NewJob("configmapJob", &testJob{Name: "configmapJob"}).When(next)
	secretJob := NewJob("secretJob", &testJob{Name: "secretJob"}).When(next)
	deploymentJob := NewJob("deploymentJob", &testJob{Name: "deploymentJob"}).When(next)
	serviceJob := NewJob("serviceJob", &testJob{Name: "serviceJob"}).When(next)
	autoscaleJob := NewJob("autoscaleJob", &testJob{Name: "autoscaleJob"}).When(next).WithResultCallback(func(result interface{}, err error) {
		pool.Close("finish")
	})
	if err := nameSpaceJob.Before(configmapJob, secretJob); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	if err := deploymentJob.After(configmapJob, secretJob); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	if err := deploymentJob.Before(serviceJob, autoscaleJob); err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	pipeline, err := NewPipeline("create", nameSpaceJob, configmapJob, secretJob, deploymentJob, serviceJob, autoscaleJob)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	pool.AddPipeline(pipeline)
	time.Sleep(time.Second)
}
