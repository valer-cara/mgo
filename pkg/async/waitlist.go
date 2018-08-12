package async

import "sync"

type Waitlist struct {
	pendingResults []*Result
	mutex          *sync.Mutex
}

func NewWaitlist() *Waitlist {
	return &Waitlist{
		pendingResults: []*Result{},
		mutex:          &sync.Mutex{},
	}
}

func (w *Waitlist) Add(pendingResult *Result) {
	w.mutex.Lock()
	w.pendingResults = append(w.pendingResults, pendingResult)
	w.mutex.Unlock()
}

func (w *Waitlist) Clear() {
	w.mutex.Lock()
	w.pendingResults = []*Result{}
	w.mutex.Unlock()
}

func (w *Waitlist) IsEmpty() bool {
	w.mutex.Lock()
	empty := (len(w.pendingResults) == 0)
	w.mutex.Unlock()
	return empty
}

func (w *Waitlist) AllDone() {
	w.mutex.Lock()
	for _, pendingResult := range w.pendingResults {
		pendingResult.Done <- true
	}
	w.mutex.Unlock()
}

func (w *Waitlist) AllError(err error) {
	w.mutex.Lock()
	for _, pendingResult := range w.pendingResults {
		pendingResult.Err <- err
	}
	w.mutex.Unlock()
}
