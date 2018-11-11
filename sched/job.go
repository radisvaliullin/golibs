package sched

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job is handler of job by schedule
type Job struct {
	// unique identifier of job
	id string

	// job time parameters (in millisecond)
	// job restart cycle
	period int
	// job running timeout
	timeout int
	// job first start delay
	startDelay int

	// state mutex
	stateMux  sync.Mutex
	isStarted bool
	isRun     bool
	// stop signal
	stop chan struct{}
	// context of job
	ctx       context.Context
	ctxCancel context.CancelFunc

	// function for execution in job
	do func(context.Context) error
}

// NewJob inits new job
func NewJob(id string, period, timeout, delay int, do func(context.Context) error) *Job {
	j := &Job{
		id:         id,
		period:     period,
		timeout:    timeout,
		startDelay: delay,
		stop:       make(chan struct{}, 1),
		do:         do,
	}
	return j
}

// Start starts job
func (j *Job) Start(ctx context.Context) {
	j.stateMux.Lock()
	defer j.stateMux.Unlock()
	if j.isStarted {
		return
	}
	j.isStarted = true
	j.ctx, j.ctxCancel = context.WithCancel(ctx)
	go j.run()
}

// Stop stops execution of job
func (j *Job) Stop() {
	j.stateMux.Lock()
	defer j.stateMux.Unlock()
	if !j.isStarted {
		return
	}
	j.stop <- struct{}{}
	j.isStarted = false
}

// Cancel breaks execution of job
func (j *Job) Cancel() {
	j.stateMux.Lock()
	defer j.stateMux.Unlock()
	if !j.isStarted {
		return
	}
	j.ctxCancel()
	j.isStarted = false
}

//
func (j *Job) run() {

	// delay running
	dlTm := time.NewTimer(time.Millisecond * time.Duration(j.startDelay))
	defer dlTm.Stop()
	select {
	case <-j.ctx.Done():
		return
	case <-j.stop:
		j.ctxCancel()
		return
	case <-dlTm.C:
	}

	// periodic run func, canceled by timeout
	do := func() {
		log.Printf("job - %v, start", j.id)
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if j.timeout == 0 {
			ctx, cancel = context.WithCancel(j.ctx)
		} else {
			ctx, cancel = context.WithTimeout(j.ctx, time.Millisecond*time.Duration(j.timeout))
		}
		defer cancel()

		// job executes
		err := j.do(ctx)
		if err != nil {
			log.Printf("job %v running err - %v", j.id, err)
		}
	}

	// periodic run
	tk := time.NewTicker(time.Millisecond * time.Duration(j.period))
	defer tk.Stop()

	do()
	for {
		select {
		case <-tk.C:
			do()
			continue
		case <-j.ctx.Done():
			return
		case <-j.stop:
			j.ctxCancel()
			return
		}
	}
}
