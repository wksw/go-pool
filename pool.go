package gopool

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// PoolRunning the pool is running
	PoolRunning = iota
	// PoolExiting the pool is exiting
	PoolExiting
	// PoolExited the pool already exited
	PoolExited
)

const (
	// DEFAULT_POOL_CAPACITY default pool capacity
	DEFAULT_POOL_CAPACITY = 10
	// DEFAULT_WORKER_NUM the number of default workers
	DEFAULT_WORKER_NUM = 1
)

var (
	// ErrPoolExit pool exit error define
	ErrPoolExit = errors.New("Pool exit")
	// ErrPoolPanic run got panic
	ErrPoolPanic = errors.New("panic")
)

// Pool job pool
type Pool struct {
	capacity uint64
	// max active goroutine
	maxActive uint64
	// the number of running goroutine
	workers uint64
	// the number of running jobs
	runners       uint64
	jobs          chan *Job
	exitCallback  func(reason string)
	panicCallback func(r interface{}, stack []byte)
	eventCallback func(event *Event)
	eventLevel    EventLevel
	status        int
	liveTime      time.Duration
	m             sync.RWMutex
}

// NewPool get a new job pool
// the pool capacity is 2, when the pool is full
// the next jobs added into pool will be block
// entil the jobs in pool was processed
// and max 2 goutine running
func NewPool(capacity, maxActive uint64) *Pool {
	if capacity == 0 {
		capacity = DEFAULT_POOL_CAPACITY
	}
	if maxActive == 0 {
		maxActive = capacity / 2
	}
	pool := &Pool{
		maxActive: maxActive,
		jobs:      make(chan *Job, capacity),
		status:    PoolRunning,
		liveTime:  time.Minute,
	}
	pool.increaseWorker()
	return pool
}

// WithExitCallback set pool exit callback function
// when pool exit will call exitCallback function
func (p *Pool) WithExitCallback(handle func(reason string)) *Pool {
	p.exitCallback = handle
	return p
}

// WithPanicCallback set goroutine panic callback function
func (p *Pool) WithPanicCallback(handle func(r interface{}, stack []byte)) *Pool {
	p.panicCallback = handle
	return p
}

// WithEventCallback set pool event callback
func (p *Pool) WithEventCallback(level EventLevel, handle func(event *Event)) *Pool {
	p.eventLevel = level
	p.eventCallback = handle
	return p
}

// AddPipeline add a new pipeline into pool
func (p *Pool) AddPipeline(pipeline *Pipeline) error {
	topJobs, err := pipeline.getTopJobs()
	if err != nil {
		return err
	}
	return p.AddJob(topJobs...)
}

// AddJob add a new job into pipeline
func (p *Pool) AddJob(jobs ...*Job) error {
	status := p.getStatus()
	if status == PoolExiting || status == PoolExited {
		return ErrPoolExit
	}
	for _, job := range jobs {
		p.sendEvent(EventLevelDebug, fmt.Sprintf("add job '%s' into queue", job.Name))
		p.jobs <- job
	}
	return nil
}

// Status get pool status
func (p *Pool) Status() int {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.status
}

// RunningJobs get running jobs number
func (p *Pool) RunningJobs() uint64 {
	return atomic.LoadUint64(&p.runners)
}

// PenddingJobs get pendding jobs number
func (p *Pool) PenddingJobs() int {
	return len(p.jobs)
}

// Workers get running goroutine number
func (p *Pool) Workers() uint64 {
	return atomic.LoadUint64(&p.workers)
}

// Close close the pool
func (p *Pool) Close(reason string) error {
	status := p.Status()
	if status == PoolExited || status == PoolExiting {
		return ErrPoolExit
	}
	// set pool status to exiting
	p.setStatus(PoolExiting)
	// wait all job finish
	p.waitAllJobFinish()
	// close job channel
	close(p.jobs)
	// wait all worker exit
	p.waitAllWorkerExit()
	if p.exitCallback != nil {
		p.exitCallback(reason)
	}
	p.setStatus(PoolExited)
	return nil
}

// wait all running jobs finish and all pendding jobs processed
func (p *Pool) waitAllJobFinish() {
	for {
		runningJobs := p.RunningJobs()
		penddingJobs := p.PenddingJobs()
		p.sendEvent(EventLevelInfo,
			fmt.Sprintf("wait all job finish, running=%d, pendding=%d",
				runningJobs, penddingJobs))
		if runningJobs == 0 && penddingJobs == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// wait all goroutine exit
func (p *Pool) waitAllWorkerExit() {
	for {
		workers := p.Workers()
		p.sendEvent(EventLevelInfo,
			fmt.Sprintf("wait all worker exit, workers=%d", workers))
		if workers == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (p *Pool) setStatus(status int) {
	p.m.Lock()
	defer p.m.Unlock()
	p.status = status
}

func (p *Pool) getStatus() int {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.status
}

func (p *Pool) increaseWorker() {
	workerNum := atomic.AddUint64(&p.workers, 1)
	p.sendEvent(EventLevelInfo,
		fmt.Sprintf("worker '%d' started, workers=%d",
			workerNum, p.Workers()))
	go p.startWorker(workerNum)
}

// atomic delete worker
func (p *Pool) decreaseWorker(workerNum uint64) {
	workers := atomic.AddUint64(&p.workers, ^uint64(0))
	p.sendEvent(EventLevelDebug,
		fmt.Sprintf("worker '%d' exited, workers=%d",
			workerNum, workers))
}

func (p *Pool) increaseRunner(job *Job) {
	runners := atomic.AddUint64(&p.runners, 1)
	p.sendEvent(EventLevelDebug,
		fmt.Sprintf("job '%s' start, running=%d, pendding=%d, workers=%d",
			job, runners, p.PenddingJobs(), p.Workers()))
}

func (p *Pool) decreaseRunner(job *Job) {
	runners := atomic.AddUint64(&p.runners, ^uint64(0))
	p.sendEvent(EventLevelDebug,
		fmt.Sprintf("job '%s' finish, runing=%d, pendding=%d, workers=%d",
			job, runners, p.PenddingJobs(), p.Workers()))
}

func (p *Pool) startWorker(workerNum uint64) {
	var currentJob *Job

	ticker := time.NewTicker(p.liveTime)
	defer ticker.Stop()

	defer func() {
		if r := recover(); r != nil {
			p.sendEvent(EventLevelError,
				fmt.Sprintf("worker '%d' execute job '%s' panic",
					workerNum, currentJob))
			p.decreaseRunner(currentJob)
			p.decreaseWorker(workerNum)
			currentJob.setResult(nil, ErrPoolPanic)
			p.increaseWorker()
			if p.panicCallback != nil {
				p.panicCallback(r, debug.Stack())
			}
		}
	}()
	for {
		select {
		case <-ticker.C:
			if p.Workers() > DEFAULT_WORKER_NUM {
				p.decreaseWorker(workerNum)
				return
			}
			ticker.Reset(p.liveTime)
		case job, ok := <-p.jobs:
			ticker.Reset(p.liveTime)
			if !ok {
				p.decreaseWorker(workerNum)
				return
			}
			if job.GetStatus() == JobCancled {
				continue
			}
			currentJob = job
			p.increaseRunner(job)

			job.setStatus(JobRunning)

			job.setResult(job.handler.Handle())

			p.decreaseRunner(job)

			p.AddJob(job.getNextExecuteJobs()...)

			if p.Workers() < p.maxActive && p.PenddingJobs() > int(p.capacity/2) {
				p.increaseWorker()
			}
		}
	}
}

func (p *Pool) sendEvent(level EventLevel, msg string) {
	if p.eventCallback != nil {
		if p.eventLevel >= level {
			p.eventCallback(&Event{level: level, msg: msg})
		}
	}
}
