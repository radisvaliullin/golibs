package sched

import (
	"context"
	"testing"
	"time"
)

func TestJob_PeriodRun(t *testing.T) {

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
