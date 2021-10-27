package gopool

import (
	"errors"
	"fmt"
)

// Pipeline pipeline define
type Pipeline struct {
	Name       string
	Jobs       []*Job
	UniqueJobs []*Job
}

// NewPipeline get a new pipeline
func NewPipeline(name string, jobs ...*Job) (*Pipeline, error) {
	pipeline := &Pipeline{
		Name: name,
		Jobs: jobs,
	}

	return pipeline.new()
}

// Cancle cancle jobs to execute
func (p *Pipeline) Cancle() {
	for _, job := range p.UniqueJobs {
		job.setStatus(JobCancled)
	}
}

func (p *Pipeline) new() (*Pipeline, error) {
	topJobs, err := p.getTopJobs()
	if err != nil {
		return &Pipeline{}, err
	}
	if err := p.isCycleAdded(topJobs); err != nil {
		return &Pipeline{}, err
	}
	p.setUniqueJobs(topJobs)
	return p, nil
}

// isCycleAdded whether cycle added
func (p *Pipeline) isCycleAdded(topJobs []*Job) error {
	var (
		visited = make(map[*Job]int)
		valid   = true
		result  jobs
		dfs     func(job *Job)
	)
	dfs = func(job *Job) {
		visited[job] = 1
		result = append(result, job)
		for _, children := range job.childrens {
			if visited[children] == 0 {
				dfs(children)
			} else if visited[children] == 1 {
				result = append(result, children)
				valid = false
				return
			}
		}
		visited[job] = 2
	}
	for _, job := range topJobs {
		if !valid {
			break
		}
		if visited[job] == 0 {
			dfs(job)
		}
	}

	if !valid {
		return fmt.Errorf("cycle added %s", result)
	}
	return nil
}

// getTopJobs get all not parent jobs
func (p *Pipeline) getTopJobs() ([]*Job, error) {
	var topJobs []*Job
	for _, job := range p.Jobs {
		if len(job.parents) == 0 {
			topJobs = append(topJobs, job)
		}
	}
	if len(topJobs) == 0 {
		return topJobs, errors.New("no top jobs")
	}
	return topJobs, nil
}

func (p *Pipeline) setUniqueJobs(jobs []*Job) {
	var (
		uniqueMap  = make(map[*Job]bool)
		uniqueFunc func(job *Job)
	)

	uniqueFunc = func(job *Job) {
		if _, ok := uniqueMap[job]; !ok {
			uniqueMap[job] = true
			p.UniqueJobs = append(p.UniqueJobs, job)
		}
		for _, jb := range job.childrens {
			uniqueFunc(jb)
		}
	}
	for _, job := range jobs {
		uniqueFunc(job)
	}
}
