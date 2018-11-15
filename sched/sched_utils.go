package sched

import "fmt"

func jobsIDUniqValidate(js []*Job) error {

	m := map[string]struct{}{}
	for _, j := range js {
		if _, ok := m[j.id]; ok {
			return fmt.Errorf("%v, id - %v", ErrJobIDNotUniq, j.id)
		}
		m[j.id] = struct{}{}
	}
	return nil
}
