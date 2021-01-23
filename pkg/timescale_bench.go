package pkg

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type TimescaleBench struct {
	logger     *logrus.Entry
	inputFile  string
	workerPool *WorkerPool
}

func (tsb *TimescaleBench) getInputFile() (*os.File, error) {
	var (
		inputFile *os.File
		err       error
	)

	if tsb.inputFile == "-" {
		tsb.logger.Info("Using stdin as input")
		inputFile = os.Stdin
	} else {
		inputFile, err = os.Open(tsb.inputFile)
		if err != nil {
			tsb.logger.Errorf("unable to open input file: %v", tsb.inputFile)
			return nil, err
		}
	}

	return inputFile, nil
}

func (tsb *TimescaleBench) Run() error {
	tsb.logger.Info("Starting...")

	inputFile, err := tsb.getInputFile()
	if err != nil {
		return err
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			tsb.logger.Warnf("unable to close file: %v", tsb.inputFile)
		}
	}()

	queryParamChan := make(chan QueryParam)
	errChan := make(chan error)
	doneChan := make(chan struct{})

	go tsb.workerPool.Run()

	go processQueryParams(inputFile, queryParamChan, errChan, doneChan)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case queryParam := <-queryParamChan:
			tsb.workerPool.Dispatch(queryParam)
		case err := <-errChan:
			tsb.logger.Warnf("Error during parsing query param: %v", err)
		case <-doneChan:
			return nil
		case sig := <-sigChan:
			tsb.logger.Infof("Received signal %v, terminating...", sig)
		}
	}
}

func NewTimescaleBench(inputFile string, numWorkers int) *TimescaleBench {
	tsb := TimescaleBench{
		logger:     logrus.WithField("component", "TimescaleBench"),
		inputFile:  inputFile,
		workerPool: newWorkerPool(numWorkers),
	}

	return &tsb
}
