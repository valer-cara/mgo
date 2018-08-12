package async

type Result struct {
	Done chan bool
	Err  chan error
}

func NewResult() *Result {
	return &Result{
		Done: make(chan bool),
		Err:  make(chan error),
	}
}
