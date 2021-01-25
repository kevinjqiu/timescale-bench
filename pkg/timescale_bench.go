package pkg

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type TimescaleBench struct {
	logger          *logrus.Entry
	inputFile       string
	workerPool      *WorkerPool
	outputFormatter OutputFormatter
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

func (tsb *TimescaleBench) parseQueryParams(inputFile *os.File) <-chan Job {
	out := make(chan Job)

	go func() {
		scanner := bufio.NewScanner(inputFile)
		scanner.Split(bufio.ScanLines)

		scanner.Scan() // skip the header
		for scanner.Scan() {
			line := scanner.Text()
			tsb.logger.Debugf("Got line: %v", line)
			queryParam, err := parseQueryParam(line)
			if err != nil {
				tsb.logger.Warn(err)
				continue
			}
			job := newJob(queryParam)
			out <- job
		}
		close(out)
	}()

	return out
}

func (tsb *TimescaleBench) Run() error {
	tsb.logger.Info("Starting...")

	inputFile, err := tsb.getInputFile()
	if err != nil {
		return err
	}

	defer inputFile.Close()

	tsb.workerPool.StartWorkers()
	br := tsb.workerPool.ProcessJobs(tsb.parseQueryParams(inputFile))

	output, err := tsb.outputFormatter(br)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func NewTimescaleBench(inputFile string, numWorkers int, dbURL string, formatter string) (*TimescaleBench, error) {
	// TODO: validate numWorkers >= 1
	wp, err := newWorkerPool(numWorkers, dbURL)
	if err != nil {
		return nil, err
	}

	outputFormatter, ok := outputFormatters[formatter]
	if !ok {
		return nil, fmt.Errorf("unrecognized output formatter: %v", formatter)
	}

	tsb := TimescaleBench{
		logger:          logrus.WithField("component", "TimescaleBench"),
		inputFile:       inputFile,
		workerPool:      wp,
		outputFormatter: outputFormatter,
	}

	return &tsb, nil
}
