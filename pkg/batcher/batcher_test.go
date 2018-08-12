package batcher

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"testing"
	"time"
)

func incrementer(x *int, err error) func() error {
	return func() error {
		*x++
		return err
	}
}

func someJobFactory(t *testing.T) Job {
	return func() error {
		sl := time.Duration(rand.Float32()*200) * time.Millisecond
		t.Log("Just a job -", sl)
		time.Sleep(sl)
		return nil
	}
}

func someFailingJobFactory(t *testing.T) Job {
	return func() error {
		sl := time.Duration(rand.Float32()*200) * time.Millisecond
		t.Log("Just a Failing job -", sl)
		time.Sleep(sl)
		return errors.New("a failing job")
	}
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	os.Exit(m.Run())
}

func TestBasicQueueCompletesInTime(t *testing.T) {
	chanBatchDone := make(chan bool)
	chanBatchErr := make(chan error)

	b := NewBatcher(&BatcherOptions{
		Done: chanBatchDone,
		Err:  chanBatchErr,
	})

	go b.Start()

	deadline := time.Now().Add(1000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	b.Queue(someJobFactory(t), nil, nil)

	select {
	case <-chanBatchDone:
		return
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func TestHooks(t *testing.T) {
	chanBatchDone := make(chan bool)
	chanBatchErr := make(chan error)

	pre, post, preItem, postItem := 0, 0, 0, 0

	b := NewBatcher(&BatcherOptions{
		Done:      chanBatchDone,
		Err:       chanBatchErr,
		PreBatch:  incrementer(&pre, nil),
		PostBatch: incrementer(&post, nil),
		PreItem:   incrementer(&preItem, nil),
		PostItem:  incrementer(&postItem, nil),
	})

	go b.Start()

	b.Queue(someJobFactory(t), nil, nil)
	b.Queue(someJobFactory(t), nil, nil)
	b.Queue(someJobFactory(t), nil, nil)
	<-chanBatchDone

	b.Queue(someJobFactory(t), nil, nil)
	b.Queue(someJobFactory(t), nil, nil)
	<-chanBatchDone

	if pre != 2 || post != 2 {
		t.Fatalf("Expected PreBatch/PostBatch to be called exactly twice, got %d/%d times", pre, post)
	}
	if preItem != 5 || postItem != 5 {
		t.Fatalf("Expected PreItem/PostItem to be called exactly 5 times, got %d/%d times", preItem, postItem)
	}
}

func TestJobStatus(t *testing.T) {
	done, errs := 0, 0

	chanBatchDone := make(chan bool)
	chanBatchErr := make(chan error)

	done1, done2 := make(chan bool), make(chan bool)
	err1, err2 := make(chan error), make(chan error)

	b := NewBatcher(&BatcherOptions{
		Done: chanBatchDone,
		Err:  chanBatchErr,
	})

	go b.Start()

	b.Queue(someJobFactory(t), done1, err1)
	b.Queue(someFailingJobFactory(t), done2, err2)

loop:
	for {
		select {
		case <-done1:
			done++
		case <-done2:
			done++
		case <-err1:
			errs++
		case <-err2:
			errs++

		case <-chanBatchDone:
			break loop
		}
	}

	if done+errs != 2 {
		t.Fatalf("Expected results from all jobs, got done %d, errs %d", done, errs)
	}
}

func TestPreBatchHookFailureTaintsAllJobs(t *testing.T) {
	done, errs := 0, 0

	chanBatchDone := make(chan bool)
	chanBatchErr := make(chan error)

	done1, done2 := make(chan bool), make(chan bool)
	err1, err2 := make(chan error), make(chan error)

	b := NewBatcher(&BatcherOptions{
		Done: chanBatchDone,
		Err:  chanBatchErr,
		PreBatch: func() error {
			return errors.New("Bad pre hook")
		},
	})

	go b.Start()

	b.Queue(someJobFactory(t), done1, err1)
	b.Queue(someFailingJobFactory(t), done2, err2)

loop:
	for {
		select {
		case <-done1:
			done++
		case <-done2:
			done++
		case <-err1:
			errs++
		case <-err2:
			errs++

		case <-chanBatchDone:
			t.Fatalf("Batch should not signal done, it should error instead")
		case <-chanBatchErr:
			break loop
		}
	}

	if done != 0 && errs != 2 {
		t.Fatalf("Expected all jobs to fail, got done %d, errs %d", done, errs)
	}
}
