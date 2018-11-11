package sched

import (
	"context"
	"testing"
	"time"
)

func TestJob_PeriodRun(t *testing.T) {

	send := make(chan struct{}, 1)
	recv := make(chan struct{}, 1)

	f := func(ctx context.Context) error {
		<-send
		recv <- struct{}{}
		return nil
	}

	j := NewJob("id", 100, 100, 0, f)
	j.Start(context.Background())

	send <- struct{}{}
	tm0 := time.NewTimer(time.Millisecond * 150)
	tm1 := time.NewTimer(time.Millisecond * 300)
	select {
	case <-recv:
		t.Logf("send recv done")
	case <-tm0.C:
		t.Fatalf("send recv fail")
	}
	send <- struct{}{}
	select {
	case <-recv:
		t.Logf("send recv done")
	case <-tm1.C:
		t.Fatalf("send recv fail")
	}
}

func TestJob_DelayRun(t *testing.T) {

	// without delay
	send0 := make(chan struct{}, 1)
	recv0 := make(chan struct{}, 1)
	f0 := func(ctx context.Context) error {
		<-send0
		recv0 <- struct{}{}
		return nil
	}

	j0 := NewJob("id0", 100, 100, 0, f0)
	j0.Start(context.Background())

	send0 <- struct{}{}
	tm0 := time.NewTimer(time.Millisecond * 150)
	select {
	case <-recv0:
		t.Logf("send recv done")
	case <-tm0.C:
		t.Fatalf("send recv fail")
	}

	// with delay
	send1 := make(chan struct{}, 1)
	recv1 := make(chan struct{}, 1)
	f1 := func(ctx context.Context) error {
		<-send1
		recv1 <- struct{}{}
		return nil
	}

	j1 := NewJob("id1", 150, 150, 150, f1)
	j1.Start(context.Background())

	send1 <- struct{}{}
	tm1 := time.NewTimer(time.Millisecond * 100)
	select {
	case <-recv1:
		t.Fatal("with delay, send recv done before delay time")
	case <-tm1.C:
		t.Logf("send recv, delay ok")
	}
}

//
func TestJob_Cancel(t *testing.T) {

	f := func(ctx context.Context) error {
		t.Logf("job do")
		select {
		case <-time.NewTimer(time.Millisecond * 200).C:
			t.Fatalf("func done by time")
		case <-ctx.Done():
			t.Logf("func done by ctx, err - %v", ctx.Err())
		}
		return nil
	}

	j := NewJob("id", 100, 100, 0, f)

	j.Start(context.Background())
	go func() {
		time.Sleep(time.Microsecond * 100)
		j.Cancel()
	}()
	j.WG()
	t.Logf("job canceled")
}
