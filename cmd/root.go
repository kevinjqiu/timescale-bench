package cmd

import (
	"fmt"
	"github.com/kevinjqiu/timescale-assignment/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var flags struct {
	input    string
	workers  int
	logLevel string
}

func initLogging(logLevel string) {
	logrus.SetOutput(os.Stderr)
	switch strings.ToLower(logLevel) {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func NewRootCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:          "timescale-bench",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			initLogging(flags.logLevel)

			if flags.input == "" {
				return fmt.Errorf("bench input file (--input) must be specified")
			}

			tsb := pkg.NewTimescaleBench(flags.input, flags.workers)
			return tsb.Run()
		},
	}

	cmd.PersistentFlags().StringVarP(&flags.input, "input", "i", "", "bench input file (csv format)")
	cmd.PersistentFlags().IntVarP(&flags.workers, "workers", "w", 5, "number of workers")
	cmd.PersistentFlags().StringVarP(&flags.logLevel, "log-level", "l", "info", "log level")
	return &cmd
}
