package sched

import (
	"context"
	"fmt"
	"testing"
	"time"
)

//
func TestSched_AddStartJobs(t *testing.T) {

	send0 := make(chan struct{}, 1)
	send1 := make(chan struct{}, 1)
	recv0 := make(chan struct{}, 1)
	recv1 := make(chan struct{}, 1)

	f0 := func(ctx context.Context) error {
		<-send0
		recv0 <- struct{}{}
		return nil
	}
	f1 := func(ctx context.Context) error {
		<-send1
		recv1 <- struct{}{}
		return nil
	}

	j0 := NewJob(context.Background(), "id0", 100, 100, 0, f0)
	j1 := NewJob(context.Background(), "id1", 100, 100, 0, f1)

	schd, err := New([]*Job{j0, j1})
	if err != nil {
		t.Fatalf("new scheduler, err - %v", err)
	}
	schd.Start()

	send0 <- struct{}{}
	send1 <- struct{}{}

	tm0 := time.NewTimer(time.Millisecond * 100)
	tm1 := time.NewTimer(time.Millisecond * 100)
	select {
	case <-recv0:
		t.Logf("send0 recv0 done")
	case <-tm0.C:
		t.Fatalf("send0 recv0 fail")
	}
	select {
	case <-recv1:
		t.Logf("send1 recv1 done")
	case <-tm1.C:
		t.Fatalf("send1 recv1 fail")
	}

	send0 <- struct{}{}
	send1 <- struct{}{}

	tm0 = time.NewTimer(time.Millisecond * 100)
	tm1 = time.NewTimer(time.Millisecond * 100)
	select {
	case <-recv0:
		t.Logf("send0 recv0 done")
	case <-tm0.C:
		t.Fatalf("send0 recv0 fail")
	}
	select {
	case <-recv1:
		t.Logf("send1 recv1 done")
	case <-tm1.C:
		t.Fatalf("send1 recv1 fail")
	}
}

//
func TestSched_JobsErrHandle(t *testing.T) {

	f0 := func(ctx context.Context) error {
		return fmt.Errorf("test job err")
	}
	f1 := func(ctx context.Context) error {
		return fmt.Errorf("test job err")
	}

	j0 := NewJob(context.Background(), "id0", 100, 100, 0, f0)
	j1 := NewJob(context.Background(), "id1", 100, 100, 0, f1)

	schd, err := New([]*Job{j0})
	if err != nil {
		t.Fatalf("new scheduler, err - %v", err)
	}
	schd.Start()

	err = schd.AddJob(j1)
	if err != nil {
		t.Fatalf("scheduler, add job, err - %v", err)
	}
	err = schd.StartJob("id1")
	if err != nil {
		t.Fatalf("scheduler, start job by id, err - %v", err)
	}

	tm0 := time.NewTimer(time.Millisecond * 50)
	tm1 := time.NewTimer(time.Millisecond * 150)
	select {
	case <-tm0.C:
		t.Fatalf("timeout before returns error")
	case err := <-schd.Err():
		t.Logf("job done with error - %v", err)

	}
	select {
	case <-tm0.C:
		t.Fatalf("timeout before returns error")
	case err := <-schd.Err():
		t.Logf("job done with error - %v", err)

	}

	select {
	case <-tm1.C:
		t.Fatalf("timeout before returns error")
	case err := <-schd.Err():
		t.Logf("job done with error - %v", err)

	}
	select {
	case <-tm1.C:
		t.Fatalf("timeout before returns error")
	case err := <-schd.Err():
		t.Logf("job done with error - %v", err)

	}

}
