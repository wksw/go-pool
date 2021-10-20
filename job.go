package gopool

import (
	"fmt"
	"sync"
)

const (
	// JobPendding job status is pendding
	JobPendding = iota
	// JobRunning job status is running
	JobRunning
	// JobSuccess job status is success
	JobSuccess
	// JobFail job status is fail
	JobFail
	// JobCancled jobs status is cancled
	JobCancled
)

// JobHandler job need implement methods
type JobHandler interface {
	Handle() (interface{}, error)
}

// Job job define
type Job struct {
	Name           string
	handler        JobHandler
	parents        []*Job
	childrens      []*Job
	status         int
	result         interface{}
	resultCallback func(interface{}, error)
	err            error
	when           func(self *Job) bool
	m              sync.RWMutex
}

type jobs []*Job

func (j jobs) String() string {
	str := ""
	for index, job := range j {
		if index == 0 {
			str += job.Name
		} else {
			str += " -> " + job.Name
		}

	}
	return str
}

// NewJob get a new job
func NewJob(name string, handler JobHandler) *Job {
	return &Job{
		Name:    name,
		handler: handler,
		status:  JobPendding,
	}
}

// WithResultCallback result callback function
func (j *Job) WithResultCallback(handler func(interface{}, error)) *Job {
	j.resultCallback = handler
	return j
}

// When set when this job execute in pipeline
func (j *Job) When(handle func(self *Job) bool) *Job {
	j.when = handle
	return j
}

// After execute after other jobs
func (j *Job) After(jobs ...*Job) error {
	for _, job := range jobs {
		if j.Name == job.Name {
			return fmt.Errorf("job '%s' dumplicate added", job.Name)
		}
		for _, children := range j.childrens {
			if children.Name == job.Name {
				return fmt.Errorf("job '%s' dumplicate added", job.Name)
			}
		}
		job.childrens = append(job.childrens, j)
		j.parents = append(j.parents, job)
		if err := j.cycleAddedCheck(); err != nil {
			return err
		}
	}
	return nil
}

// Before execute before other jobs
func (j *Job) Before(jobs ...*Job) error {
	for _, job := range jobs {
		if j.Name == job.Name {
			return fmt.Errorf("job '%s' dumplicate added", job.Name)
		}
		for _, children := range j.childrens {
			if children.Name == job.Name {
				return fmt.Errorf("job '%s' dumplicate added", job.Name)
			}
		}
		job.parents = append(job.parents, j)
		j.childrens = append(j.childrens, job)
		if err := j.cycleAddedCheck(); err != nil {
			return err
		}
	}
	return nil
}

// GetStatus get job execute status
func (j *Job) GetStatus() int {
	j.m.RLock()
	defer j.m.RUnlock()
	return j.status
}

// GetResult get job execute result and error
func (j *Job) GetResult() (interface{}, error) {
	j.m.RLock()
	defer j.m.RUnlock()
	return j.result, j.err
}

// String job format
func (j *Job) String() string {
	return j.Name
}

// GetUpstreams get parent jobs
func (j *Job) GetUpstreams() []*Job {
	return j.parents
}

// GetDownstreams get children jobs
func (j *Job) GetDownstreams() []*Job {
	return j.childrens
}

func (j *Job) getNextExecuteJobs() []*Job {
	var downStreams []*Job
	for _, job := range j.childrens {
		if job.when != nil {
			if job.when(job) {
				downStreams = append(downStreams, job)
			}
		} else {
			downStreams = append(downStreams, job)
		}
	}
	return downStreams
}

func (j *Job) setResult(result interface{}, err error) {
	j.m.Lock()
	defer j.m.Unlock()
	j.result = result
	j.err = err
	if err != nil {
		j.status = JobFail
	} else {
		j.status = JobSuccess
	}
	if j.resultCallback != nil {
		j.resultCallback(result, err)
	}
}

func (j *Job) setStatus(status int) {
	j.m.Lock()
	defer j.m.Unlock()
	j.status = status
}

func (j *Job) cycleAddedCheck() error {
	var (
		visited = make(map[*Job]int)
		valid   = true
		result  jobs
		dfs     func(job *Job)
	)
	dfs = func(job *Job) {
		visited[job] = 1
		for _, children := range j.childrens {
			if visited[children] == 0 {
				dfs(children)
			} else if visited[children] == 1 {
				result = append(result, children)
				valid = false
				return
			}
			visited[job] = 2
			result = append(result, job)
		}
	}
	dfs(j)
	if !valid {
		return fmt.Errorf("cycle added %s", result)
	}
	return nil
}
