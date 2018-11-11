package sched

import (
	"context"
	"log"
	"time"
)

// Job is handler of job by schedule
type Job struct {
	// unique identifier of job
	id string

	// job time parameters (in sec)
	// job restart cycle
	period int
	// job running timeout
	timeout int
	// job first start delay
	startDelay int

	// context of job
	ctx       context.Context
	ctxCancel context.CancelFunc

	// function for execution in job
	do func(context.Context) error
}

// NewJob inits new job
func NewJob(ctx context.Context, id string, period, timeout, delay int, do func(context.Context) error) *Job {
	ctx, cancel := context.WithCancel(ctx)
	j := &Job{
		id:         id,
		period:     period,
		timeout:    timeout,
		startDelay: delay,
		ctx:        ctx,
		ctxCancel:  cancel,
		do:         do,
	}
	return j
}

// Start starts job
func (j *Job) Start() {

	go j.run()
}

// Stop breaks execution of job
func (j *Job) Stop() {
	j.ctxCancel()
}

//
func (j *Job) run() {

	// delay running
	dlTm := time.NewTimer(time.Second * time.Duration(j.startDelay))
	defer dlTm.Stop()
	select {
	case <-j.ctx.Done():
		return
	case <-dlTm.C:
	}

	// periodic run func, canceled by timeout
	do := func() {
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if j.timeout == 0 {
			ctx, cancel = context.WithCancel(j.ctx)
		} else {
			ctx, cancel = context.WithTimeout(j.ctx, time.Second*time.Duration(j.timeout))
		}
		defer cancel()

		// job executes
		err := j.do(ctx)
		if err != nil {
			log.Printf("job %v running err - %v", j.id, err)
		}
	}

	// periodic run
	tk := time.NewTicker(time.Second * time.Duration(j.period))
	defer tk.Stop()

	do()
	for {
		select {
		case <-tk.C:
			log.Printf("job - %v, start", j.id)
			do()
			continue
		case <-j.ctx.Done():
			return
		}
	}
}
