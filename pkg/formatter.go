package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var outputFormatters = map[string]OutputFormatter{
	"human": humanReadableOutputFormatter,
	"json": jsonOutputFormatter,
}

type OutputFormatter func(BenchmarkResult) (string, error)

func humanReadableOutputFormatter(br BenchmarkResult) (string, error) {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Num Queries: %d\n", br.NumQueries))
	buffer.WriteString(fmt.Sprintf("Num Errors: %d\n", br.NumErrors))
	buffer.WriteString(fmt.Sprintf("Total Processing time: %v\n", br.TotalProcessingTime))
	buffer.WriteString(fmt.Sprintf("Min time: %v\n", br.Min))
	buffer.WriteString(fmt.Sprintf("Max time: %v\n", br.Max))
	buffer.WriteString(fmt.Sprintf("Average time: %v\n", br.Average))
	buffer.WriteString(fmt.Sprintf("Median time: %v\n", br.Median))

	return buffer.String(), nil

}

func jsonOutputFormatter(br BenchmarkResult) (string, error) {
	b, err := json.MarshalIndent(br, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}