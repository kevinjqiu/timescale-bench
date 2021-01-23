package main

import (
	"github.com/kevinjqiu/timescale-assignment/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	command := cmd.NewRootCommand()
	if err := command.Execute(); err != nil {
		logrus.Error(err)
	}
}