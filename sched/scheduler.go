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
}

// New inits new scheduler
// js - list of jobs for schedule
func New(js []*Job) (*Sched, error) {
	if js == nil {
		js = []*Job{}
	}
	err := jobsIDUniqValidate(js)
	if err != nil {
		return nil, err
	}
	s := &Sched{
		jobs: js,
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
	return nil
}

// StartJob stops job by id
func (s *Sched) StartJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	exist := false
	for _, jb := range s.jobs {
		if jb.id == id {
			exist = true
			jb.Start(nil)
			return nil
		}
	}
	if !exist {
		return fmt.Errorf("%v: start job, job with id %v doesn't exist", pkgPref, id)
	}
	return nil
}

// StopJob stops job by id
func (s *Sched) StopJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	exist := false
	for _, jb := range s.jobs {
		if jb.id == id {
			exist = true
			jb.Stop()
			return nil
		}
	}
	if !exist {
		return fmt.Errorf("%v: stop job, job with id %v doesn't exist", pkgPref, id)
	}
	return nil
}

// CancelJob cancels job by id
func (s *Sched) CancelJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	exist := false
	for _, jb := range s.jobs {
		if jb.id == id {
			exist = true
			jb.Cancel()
			return nil
		}
	}
	if !exist {
		return fmt.Errorf("%v: cancel job, job with id %v doesn't exist", pkgPref, id)
	}
	return nil
}

// DelJob deletes job by id (job must be stopped or canceled)
func (s *Sched) DelJob(id string) error {
	s.jsMux.Lock()
	defer s.jsMux.Unlock()
	exist := false
	for i, jb := range s.jobs {
		if jb.id == id {
			exist = true
			copy(s.jobs[i:], s.jobs[i+1:])
			s.jobs[len(s.jobs)-1] = nil
			s.jobs = s.jobs[:len(s.jobs)-1]
			return nil
		}
	}
	if !exist {
		return fmt.Errorf("%v: delete job, job with id %v doesn't exist", pkgPref, id)
	}
	return nil
}
