/*
 * Batcher: can be used to queue multiple work items (functions)
 * Processing starts automatically.
 * If additional work is queued before the current one is finished, the Done()
 * signal is postponed until everything is complete.
 */
package batcher

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const defaultMaxQueueSize = 20

type Batcher struct {
	queue      chan *jobSpec
	options    *BatcherOptions
	mutex      *sync.Mutex
	processing bool
}

type BatcherOptions struct {
	// Pre,Post batch hooks
	PreBatch  func() error
	PostBatch func() error

	// Pre,Post job hooks
	// Job execution stops if PreHook fails
	// PostHook is not executed if PreHook or job fails
	PreItem  func() error
	PostItem func() error

	// Done, Err chans
	Done chan bool
	Err  chan error

	// Limit batch size
	MaxQueueSize int

	// Whether to continue processing the batch on errors
	// Pre/Post hooks do not stop execution, regardless of this flag (TODO?)
	ContinueOnErrors bool
}

type Job func() error

type jobSpec struct {
	job  Job
	done chan bool
	err  chan error
}

func NewBatcher(options *BatcherOptions) *Batcher {
	maxQueueSize := options.MaxQueueSize
	if maxQueueSize == 0 {
		maxQueueSize = defaultMaxQueueSize
	}

	return &Batcher{
		options:    options,
		queue:      make(chan *jobSpec, maxQueueSize),
		mutex:      &sync.Mutex{},
		processing: false,
	}
}

// Start a loop to watch for new elements in queue and process them
// Usage: `go mybatcher.Start()`
func (b *Batcher) Start() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		// TODO: Maybe something else should trigger batch processing instead of this Ticker
		case <-ticker.C:
			if len(b.queue) > 0 {
				log.Debugln("Processing batch...")
				b.process()
			}
		}
	}
}

// Queue a new item to be batch-processed
func (b *Batcher) Queue(job Job, chanDone chan bool, chanErr chan error) {
	j := jobSpec{
		job:  job,
		done: chanDone,
		err:  chanErr,
	}

	b.queue <- &j
}

func (b *Batcher) process() {
	// When an error occurs, mark batch as tainted and signal an error on all
	// jobs in batch
	batchTainted := false

	statsProcessed := 0

	if b.options.PreBatch != nil {
		if err := b.options.PreBatch(); err != nil {
			b.signalBatchError(err)
			batchTainted = true
		}
	}

loop:
	for {
		select {
		case itemJobSpec := <-b.queue:
			statsProcessed++
			if batchTainted {
				itemJobSpec.signalJobError(errors.New("Cannot process request. An error occurred."))
				continue
			}

			if err := b.processJob(itemJobSpec); err != nil {
				log.Warnln("Failed batched job: ", err)
			}
		default:
			if batchTainted {
				// No more postBatch hook execution, no more signalBatchDone
				return
			} else {
				log.Debugf("Done all work (%d jobs)", statsProcessed)
				break loop
			}
		}
	}

	if b.options.PostBatch != nil {
		if err := b.options.PostBatch(); err != nil {
			b.signalBatchError(err)
		}
	}

	b.signalBatchDone()
}

func (b *Batcher) processJob(js *jobSpec) error {
	if b.options.PreItem != nil {
		if err := b.options.PreItem(); err != nil {
			js.signalJobError(err)
			return err
		}
	}

	if err := js.job(); err != nil {
		js.signalJobError(err)
		return err
	}

	if b.options.PostItem != nil {
		if err := b.options.PostItem(); err != nil {
			js.signalJobError(err)
			return err
		}
	}

	js.signalJobDone()
	return nil
}

func (j *jobSpec) signalJobDone() {
	if j.done != nil {
		j.done <- true
	}
}
func (j *jobSpec) signalJobError(err error) {
	if j.err != nil {
		j.err <- err
	}
}

func (b *Batcher) signalBatchDone() {
	if b.options.Done != nil {
		b.options.Done <- true
	}
}

func (b *Batcher) signalBatchError(err error) {
	if b.options.Err != nil {
		b.options.Err <- err
	}
}
