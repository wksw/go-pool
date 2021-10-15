package main

import (
	"fmt"
	"log"
	"time"

	gopool "github.com/wksw/go-pool"
)

type jobA struct {
	Name string
}

type jobB struct {
	Name string
}

type jobC struct {
	Name string
}

type jobD struct {
	Name string
}

var _ gopool.JobHandler = &jobA{}
var _ gopool.JobHandler = &jobB{}
var _ gopool.JobHandler = &jobC{}
var _ gopool.JobHandler = &jobD{}

func (j *jobA) Handle() (interface{}, error) {
	fmt.Println("jobA handler")
	return nil, nil
}

func (j *jobB) Handle() (interface{}, error) {
	fmt.Println("jobB handler")
	return nil, nil
}

func (j *jobC) Handle() (interface{}, error) {
	fmt.Println("jobC handler")
	return nil, nil
}
func (j *jobD) Handle() (interface{}, error) {
	fmt.Println("jobD handler")
	return nil, nil
}

func main() {

	jobA := gopool.NewJob("jobA", &jobA{Name: "jobA"})
	jobB := gopool.NewJob("jobB", &jobB{Name: "jobB"})
	jobC := gopool.NewJob("jobC", &jobC{Name: "jobC"})
	jobD := gopool.NewJob("jobD", &jobD{Name: "jobD"})

	if err := jobC.When(func(self *gopool.Job) bool {
		for _, job := range self.GetUpstreams() {
			if job.GetStatus() != gopool.JobSuccess {
				return false
			}
		}
		return true
	}).After(jobA, jobB); err != nil {
		log.Fatal("A, B->C ", err.Error())
	}

	if err := jobD.After(jobA, jobB, jobC); err != nil {
		log.Fatal("A, B, C->D ", err.Error())
	}

	pool := gopool.NewPool(2, 2)

	pipeline, err := gopool.NewPipeline("pipeline", jobA, jobB, jobC, jobD)
	if err != nil {
		log.Fatal(err.Error())
	}

	pool.AddPipeline(pipeline)

	time.Sleep(10 * time.Second)

	// pool.Close("finish")
}
