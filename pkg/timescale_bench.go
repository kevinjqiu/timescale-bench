package pkg

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"os"
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

	doneChan := make(chan struct{})

	go tsb.workerPool.Run(doneChan)

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	scanner.Scan() // skip the header
	for scanner.Scan() {
		line := scanner.Text()
		logrus.Debugf("Got line: %v", line)
		queryParam, err := parseQueryParam(line)
		if err != nil {
			return err
		}
		tsb.workerPool.dispatch(queryParam)
	}

	doneChan <- struct{}{}
	<- doneChan
	return nil
}

func NewTimescaleBench(inputFile string, numWorkers int) (*TimescaleBench, error) {
	wp, err := newWorkerPool(numWorkers)
	if err != nil {
		return nil, err
	}

	tsb := TimescaleBench{
		logger:     logrus.WithField("component", "TimescaleBench"),
		inputFile:  inputFile,
		workerPool: wp,
	}

	return &tsb, nil
}
