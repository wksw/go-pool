package main

import (
	"fmt"
	"log"
	"time"

	gopool "github.com/wksw/go-pool"
)

type jobA struct {
	Name string
	A    time.Duration
}

type jobB struct {
	Name string
	B    int
}

type jobC struct {
	Name string
	C    int64
}

type jobD struct {
	Name string
	D    string
}

type jobE struct {
	Name string
	E    string
}

var _ gopool.JobHandler = &jobA{}
var _ gopool.JobHandler = &jobB{}
var _ gopool.JobHandler = &jobC{}
var _ gopool.JobHandler = &jobD{}
var _ gopool.JobHandler = &jobE{}

func (j *jobA) Handle() (interface{}, error) {
	time.Sleep(j.A)
	fmt.Println(j.Name, "花了", j.A.Seconds(), "秒", time.Now().Unix())
	return nil, nil
}

func (j *jobB) Handle() (interface{}, error) {
	time.Sleep(time.Duration(j.B) * time.Second)
	fmt.Println(j.Name, "花了", j.B, "秒", time.Now().Unix())
	return nil, nil
}

func (j *jobC) Handle() (interface{}, error) {
	time.Sleep(time.Duration(j.C) * time.Second)
	fmt.Println(j.Name, "花了", j.C, "秒", time.Now().Unix())
	return nil, nil
}

func (j *jobD) Handle() (interface{}, error) {
	fmt.Println(j.Name, j.D, time.Now().Unix())
	return nil, nil
}

func (j *jobE) Handle() (interface{}, error) {
	fmt.Println(j.Name, j.E, time.Now().Unix())
	return nil, nil
}

func main() {

	jobA := gopool.NewJob("起床", &jobA{Name: "起床", A: time.Second})
	jobB := gopool.NewJob("洗脸", &jobB{Name: "洗脸", B: 2})
	jobC := gopool.NewJob("刷牙", &jobC{Name: "刷牙", C: 3})
	jobD := gopool.NewJob("深呼吸", &jobD{Name: "深呼吸", D: "一大口"})
	jobE := gopool.NewJob("上班", &jobE{Name: "上班", E: ""})

	if err := jobB.When(func(self *gopool.Job) bool {
		for _, job := range self.GetUpstreams() {
			if job.GetStatus() != gopool.JobSuccess {
				return false
			}
		}
		return true
	}).After(jobA); err != nil {
		log.Fatal("A -> B ", err.Error())
	}

	if err := jobC.When(func(self *gopool.Job) bool {
		for _, job := range self.GetUpstreams() {
			if job.GetStatus() != gopool.JobSuccess {
				return false
			}
		}
		return true
	}).After(jobA); err != nil {
		log.Fatal("A -> C ", err.Error())
	}

	if err := jobE.When(func(self *gopool.Job) bool {
		for _, job := range self.GetUpstreams() {
			if job.GetStatus() != gopool.JobSuccess {
				return false
			}
		}
		return true
	}).After(jobB, jobC); err != nil {
		log.Fatal("A, B, C, D -> E ", err.Error())
	}

	if err := jobD.After(jobA, jobB, jobC, jobE); err != nil {
		log.Fatal("A, B, C, E->D ", err.Error())
	}

	pool := gopool.NewPool(10, 5).
		WithEventCallback(gopool.EventLevelDebug, func(event *gopool.Event) {
			// fmt.Println(event)
		}).
		WithPanicCallback(func(r interface{}) {
			fmt.Println("panic", r)
		}).
		WithExitCallback(func(reason string) {
			fmt.Println("pool exist beacuse", reason)
		})

	pipeline, err := gopool.NewPipeline("pipeline", jobA, jobB, jobC, jobD)
	if err != nil {
		log.Fatal(err.Error())
	}

	pool.AddPipeline(pipeline)

	// 起床 1 秒
	// 洗脸 2 秒
	// 刷牙 3 秒
	// 上班 1 秒
	// 先起床， 然后同时洗脸刷牙，按刷牙最大耗时计算, 共 3+1=4秒
	// 如果在4秒钟之前pool exit 则出不了门
	// 否则没穿衣服就出门了

	time.Sleep(2 * time.Second)
	pipeline.Cancle()

	pool.Close("没穿衣服")
}
