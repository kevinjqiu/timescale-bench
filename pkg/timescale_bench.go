package pkg

import (
	"github.com/sirupsen/logrus"
	"os"
)

type TimescaleBench struct {
	inputFile string
	workers   int
}

func (tsb *TimescaleBench) getInputFile() (*os.File, error) {
	var (
		inputFile *os.File
		err       error
	)

	if tsb.inputFile == "-" {
		logrus.Info("Using stdin as input")
		inputFile = os.Stdin
	} else {
		inputFile, err = os.Open(tsb.inputFile)
		if err != nil {
			logrus.Errorf("unable to open input file: %v", tsb.inputFile)
			return nil, err
		}
	}

	return inputFile, nil
}

func (tsb *TimescaleBench) Run() error {
	inputFile, err := tsb.getInputFile()
	if err != nil {
		return err
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			logrus.Warn("unable to close file: %v", tsb.inputFile)
		}
	}()


	queryParamChan := make(chan QueryParam)
	errChan := make(chan error)

	processQueryParams(inputFile, queryParamChan, errChan)

	return nil
}

func NewTimescaleBench(inputFile string, workers int) *TimescaleBench {
	tsb := TimescaleBench{
		inputFile: inputFile,
		workers:   workers,
	}

	return &tsb
}
