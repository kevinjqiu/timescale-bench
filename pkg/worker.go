package pkg

type Worker struct {
	inputChan     chan<- QueryParam
	errChan       <-chan error
	terminateChan chan<- struct{}
}

func newWorker(inputChan chan<- QueryParam, errChan <-chan error, terminateChan chan<- struct{}) *Worker {
	return &Worker{
		inputChan:     inputChan,
		errChan:       errChan,
		terminateChan: terminateChan,
	}
}
