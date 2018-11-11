package sched

import "fmt"

// Errors
var (
	ErrJobIDNotUniq = fmt.Errorf("%v: job id is not unique", pkgPref)
)
