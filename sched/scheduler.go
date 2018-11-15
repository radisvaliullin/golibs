package sched

import (
	"fmt"
	"sync"
)

const (
	pkgPref = "scheduler"
)

// Sched is scheduler of jobs
type Sched struct {
	jsMux sync.Mutex
	jobs  []*Job

	//
	err   chan error
	errWG sync.WaitGroup
}

// New inits new scheduler
// js - list of jobs for schedule
func New(js []*Job) (*Sched, error) {
	if js == nil {
		js = []*Job{}
	}
	// validate
	err := jobsIDUniqValidate(js)
	if err != nil {
		return nil, err
	}
	// scheduler
	s := &Sched{
		jobs: js,
		err:  make(chan error, 100),
	}
	// run job errors handler
	for _, j := range js {
		s.errWG.Add(1)
		go s.joinJobErr(j.Err())
	}
	return s, nil
}

// Start starts scheduler
func (s *Sched) Start() {
	s.jsMux.Lock()
	for _, j := range s.jobs {
		j.Start(nil)
	}
	s.jsMux.Unlock()
}

// AddJob adds new job to pool
func (s *Sched) AddJob(j *Job) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	for _, jb := range s.jobs {
		if jb.id == j.id {
			return fmt.Errorf("%v, id - %v", ErrJobIDNotUniq, j.id)
		}
	}
	s.jobs = append(s.jobs, j)
	// run job errors handler
	s.errWG.Add(1)
	go s.joinJobErr(j.Err())
	return nil
}

// StartJob stops job by id
func (s *Sched) StartJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	for _, jb := range s.jobs {
		if jb.id == id {
			jb.Start(nil)
			return nil
		}
	}
	return fmt.Errorf("%v: start job, job with id %v doesn't exist", pkgPref, id)
}

// StopJob stops job by id
func (s *Sched) StopJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	for _, jb := range s.jobs {
		if jb.id == id {
			jb.Stop()
			return nil
		}
	}
	return fmt.Errorf("%v: stop job, job with id %v doesn't exist", pkgPref, id)
}

// CancelJob cancels job by id
func (s *Sched) CancelJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	for _, jb := range s.jobs {
		if jb.id == id {
			jb.Cancel()
			return nil
		}
	}
	return fmt.Errorf("%v: cancel job, job with id %v doesn't exist", pkgPref, id)
}

// JobWG job's wait group (blocking)
func (s *Sched) JobWG(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	for _, jb := range s.jobs {
		if jb.id == id {
			jb.WG()
			return nil
		}
	}
	return fmt.Errorf("%v: delete job, job with id %v doesn't exist", pkgPref, id)
}

// DelJob deletes job by id (job must be stopped or canceled)
func (s *Sched) DelJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	for i, jb := range s.jobs {
		if jb.id == id {
			if jb.isStarted || jb.isRun {
				return fmt.Errorf("%v: delete job, job with id %v can't be started or runned", pkgPref, id)
			}
			jb.ErrClose()
			copy(s.jobs[i:], s.jobs[i+1:])
			s.jobs[len(s.jobs)-1] = nil
			s.jobs = s.jobs[:len(s.jobs)-1]
			return nil
		}
	}
	return fmt.Errorf("%v: delete job, job with id %v doesn't exist", pkgPref, id)
}

// Err returns error chan
func (s *Sched) Err() <-chan error {
	return s.err
}

// WG scheduler jobs error handler wait group (blocking)
func (s *Sched) WG() {
	s.errWG.Wait()
}

// ErrClose closes error
func (s *Sched) ErrClose() <-chan error {
	close(s.err)
	return s.err
}

//
func (s *Sched) joinJobErr(jerr <-chan error) {
	defer s.errWG.Done()
	for e := range jerr {
		s.err <- e
	}
}
