package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/radisvaliullin/golibs/sched"
)

func main() {

	send0 := make(chan string, 1)
	sendErr0 := make(chan error, 1)

	f0 := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return fmt.Errorf("func 0 terminated by context")
		case msg := <-send0:
			log.Printf("func 0, msg - %v", msg)
		case err := <-sendErr0:
			return err
		}
		return nil
	}

	j0 := sched.NewJob(context.Background(), "id0", 500, 550, 0, f0)

	s, err := sched.New([]*sched.Job{j0})
	if err != nil {
		log.Fatalf("new scheduler err - %v", err)
	}
	s.Start()

	errHandlerStopped := make(chan struct{})
	go func() {
		for err := range s.Err() {
			log.Printf("scheduler: err - %v", err)
		}
		errHandlerStopped <- struct{}{}
	}()

	// send msg
	for i := 0; i < 5; i++ {
		tm := time.NewTimer(time.Millisecond * 500)
		select {
		case <-tm.C:
			send0 <- fmt.Sprintf("1000 ms timer, tick - %v", i)
		}
	}

	// send err
	for i := 0; i < 5; i++ {
		tm := time.NewTimer(time.Millisecond * 500)
		select {
		case <-tm.C:
			sendErr0 <- fmt.Errorf("error: 1000 ms timer, tick - %v", i)
		}
	}

	// timeout
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 550)
	}

	// send msg
	for i := 0; i < 5; i++ {
		tm := time.NewTimer(time.Millisecond * 500)
		select {
		case <-tm.C:
			send0 <- fmt.Sprintf("1000 ms timer, tick - %v", i)
		}
	}

	log.Print("stop job id0")
	err = s.StopJob("id0")
	if err != nil {
		log.Fatalf("scheduler: stop job, id - %v, err - %v", "id0", err)
	}
	err = s.JobWG("id0")
	if err != nil {
		log.Fatalf("scheduler: job wg, id - %v, err - %v", "id0", err)
	}
	err = s.DelJob("id0")
	if err != nil {
		log.Fatalf("scheduler: delete job, id - %v, err - %v", "id0", err)
	}
	s.WG()
	s.ErrClose()

	<-errHandlerStopped
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 500)
		log.Print("End")
	}
}
