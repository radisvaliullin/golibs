package sched

import "fmt"

// Errors
var (
	ErrJobIDNotUniq = fmt.Errorf("%v: job id is not unique", pkgPref)
)

// JobErr job's error wrapper
type JobErr struct {
	id  string
	err error
}

// Error interface method
func (e JobErr) Error() string {
	return fmt.Sprintf("job %v: %v", e.id, e.err)
}

// ID returns job err id
func (e *JobErr) ID() string {
	return e.id
}

// Err returns job original err
func (e *JobErr) Err() error {
	return e.err
}
