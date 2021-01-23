# timescale-bench

Benchmark timescaledb `select` performance.

## Usage

    timescale-bench --input query_params.csv --workers 10

### Flags

* `--input` - path to the query params csv file. Use `-` to read the input from STDIN.
* `--workers` - number of concurrent workers
