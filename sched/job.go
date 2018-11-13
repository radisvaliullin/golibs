package sched

import (
	"context"
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
	stateMux sync.Mutex
	// start/stop trigger
	isStarted bool
	// handler statuse
	isRun bool
	// job active statuse
	isDo bool
	// stop signal
	stop chan struct{}
	// wait group
	wg sync.WaitGroup
	// context of job
	baseCtx   context.Context
	ctx       context.Context
	ctxCancel context.CancelFunc

	// function for execution in job
	do func(context.Context) error

	// error
	err chan error
}

// NewJob inits new job
func NewJob(ctx context.Context, id string, period, timeout, delay int, do func(context.Context) error) *Job {
	j := &Job{
		id:         id,
		period:     period,
		timeout:    timeout,
		startDelay: delay,
		stop:       make(chan struct{}, 1),
		baseCtx:    ctx,
		do:         do,
		err:        make(chan error, 10),
	}
	return j
}

// Start starts job
func (j *Job) Start(ctx context.Context) {
	j.stateMux.Lock()
	defer j.stateMux.Unlock()
	if j.isStarted || j.isRun {
		return
	}
	j.isStarted = true
	// don't forget call cancel
	if ctx != nil {
		j.ctx, j.ctxCancel = context.WithCancel(ctx)
	} else if j.baseCtx != nil {
		j.ctx, j.ctxCancel = context.WithCancel(j.baseCtx)
	} else {
		j.ctx, j.ctxCancel = context.WithCancel(context.Background())
	}
	j.wg.Add(1)
	go j.run()
}

// Stop stops execution of job (finished current cycle)
func (j *Job) Stop() {
	j.stateMux.Lock()
	defer j.stateMux.Unlock()
	if !j.isStarted {
		return
	}
	j.stop <- struct{}{}
	j.isStarted = false
}

// Cancel breaks execution of job (break current cycle)
func (j *Job) Cancel() {
	j.stateMux.Lock()
	defer j.stateMux.Unlock()
	if !j.isStarted {
		return
	}
	j.ctxCancel()
	j.isStarted = false
}

// WG waits jop stopping (blocking)
func (j *Job) WG() {
	j.wg.Wait()
}

// IsRun checks job handler state
func (j *Job) IsRun() bool {
	var st bool
	j.stateMux.Lock()
	st = j.isRun
	j.stateMux.Unlock()
	return st
}

// IsDo checks job execution state
func (j *Job) IsDo() bool {
	var st bool
	j.stateMux.Lock()
	st = j.isDo
	j.stateMux.Unlock()
	return st
}

// Err errors chan
func (j *Job) Err() <-chan error {
	return j.err
}

//
func (j *Job) run() {
	defer j.wg.Done()
	defer func() {
		j.stateMux.Lock()
		j.isRun = false
		j.stateMux.Unlock()
	}()
	j.stateMux.Lock()
	j.isRun = true
	j.stateMux.Unlock()

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
		defer func() {
			j.stateMux.Lock()
			j.isDo = false
			j.stateMux.Unlock()
		}()
		j.stateMux.Lock()
		j.isDo = true
		j.stateMux.Unlock()

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
			j.err <- JobErr{id: j.id, err: err}
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
