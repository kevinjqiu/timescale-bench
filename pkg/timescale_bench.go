package pkg

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"os"
	"time"
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

	resultChan := make(chan time.Duration)

	tsb.workerPool.startWorkers(resultChan)

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	scanner.Scan() // skip the header
	for scanner.Scan() {
		line := scanner.Text()
		tsb.logger.Debugf("Got line: %v", line)
		queryParam, err := parseQueryParam(line)
		if err != nil {
			return err
		}
		tsb.workerPool.dispatch(queryParam)
	}

	results := make([]time.Duration, 0, 0)

	for {
		select {
		case result := <- resultChan:
			results = append(results, result)
		}
	}

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
