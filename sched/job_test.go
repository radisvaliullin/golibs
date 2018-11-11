package sched

import (
	"context"
	"testing"
	"time"
)

func TestJob_Run(t *testing.T) {

	send := make(chan struct{})
	recv := make(chan struct{})

	f := func(ctx context.Context) error {
		<-send
		recv <- struct{}{}
		return nil
	}

	j := NewJob("id", 100, 100, 0, f)

	j.Start(context.Background())

	send <- struct{}{}
	select {
	case <-recv:
		t.Logf("send recv done")
	case <-time.NewTimer(time.Millisecond * 200).C:
		t.Fatalf("send recv fail")
	}
	send <- struct{}{}
	select {
	case <-recv:
		t.Logf("send recv done")
	case <-time.NewTimer(time.Millisecond * 200).C:
		t.Fatalf("send recv fail")
	}
}
