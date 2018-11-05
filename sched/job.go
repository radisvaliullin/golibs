package sched

import (
	"context"
	"log"
	"time"
)

// Job is schedule job
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

// NewJob -
func NewJob(ctx context.Context, n string, per, timeout, delay int, do func(context.Context) error) *Job {
	ctx, cancel := context.WithCancel(ctx)
	j := &Job{
		id:         n,
		period:     per,
		timeout:    timeout,
		startDelay: delay,
		ctx:        ctx,
		ctxCancel:  cancel,
		do:         do,
	}
	return j
}

// Start -
func (j *Job) Start() {

	go j.run()
}

//
func (j *Job) run() {

	// delay running
	dlTk := time.NewTicker(time.Second * time.Duration(j.startDelay))
	defer dlTk.Stop()
	select {
	case <-j.ctx.Done():
		return
	case <-dlTk.C:
	}

	// periodic run func
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
