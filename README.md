# timescale-bench

Benchmark timescaledb `select` performance.

## Usage

    timescale-bench --input query_params.csv --workers 10

### Flags

```
Usage:
  timescale-bench [flags]

Flags:
  -d, --db-url string             the timescaledb url to benchmark against (default "postgres://postgres:password@localhost:5432/homework")
  -h, --help                      help for timescale-bench
  -i, --input string              bench input file in csv format
  -l, --log-level string          log level (default "info")
  -o, --output-formatter string   output formatter (default "human")
  -w, --workers int               number of workers (default 5)
```

## Steps to run on test dataset
1. Download the test data
   
```
make data
```
2. Unzip
```
mkdir data/ && mv data.tar.gz data/ && cd data && tar -xzvf data.tar.gz
```
3. Start timescaledb using docker

```
make start-timescale
```

4. Initialize dataset (including setting up the database schema and importing the dataset)

```
make init
```

5. Compile the binary

```
make build
```

6. Run it

```
bin/timescale-bench --input data/query_params.csv -l warn
```

```
→ bin/timescale-bench --input data/query_params.csv -l warn
Num Queries: 200
Num Errors: 0
Total Processing time: 1.1858086s
Min time: 5.929043ms
Max time: 5.929043ms
Average time: 5.929043ms
Median time: 5.929043ms
```

Alternatively, use `-o json` to have the metrics displayed in json format (for easier parsing by other tools):

```
→ bin/timescale-bench --input data/query_params.csv -l warn -o json
{
  "NumQueries": 200,
  "NumErrors": 0,
  "TotalProcessingTime": 1071775600,
  "Min": 5358878,
  "Max": 5358878,
  "Average": 5358878,
  "Median": 5358878
}
```

Use `-l/--log-level` to adjust the log level.

e.g.,:
```
→ bin/timescale-bench --input data/query_params.csv -l info
INFO[0000] Starting...                                   component=TimescaleBench
INFO[0000] Start the worker pool                         component=WorkerPool
INFO[0000] Running worker <Worker: 4>                    component=worker-4
INFO[0000] Running worker <Worker: 0>                    component=worker-0
INFO[0000] Running worker <Worker: 2>                    component=worker-2
INFO[0000] Running worker <Worker: 3>                    component=worker-3
INFO[0000] Running worker <Worker: 1>                    component=worker-1
INFO[0000] Finished dispatching all jobs                 component=WorkerPool
INFO[0000] Finished processing jobs                      component=worker-2
INFO[0000] Close database connection                     component=worker-2
INFO[0000] Finished processing jobs                      component=worker-0
INFO[0000] Close database connection                     component=worker-0
INFO[0000] Finished processing jobs                      component=worker-4
INFO[0000] Close database connection                     component=worker-4
INFO[0000] Finished processing jobs                      component=worker-3
INFO[0000] Close database connection                     component=worker-3
INFO[0000] Finished processing jobs                      component=worker-1
INFO[0000] Close database connection                     component=worker-1
Num Queries: 200
Num Errors: 0
Total Processing time: 1.0919558s
Min time: 5.459779ms
Max time: 5.459779ms
Average time: 5.459779ms
Median time: 5.459779ms
```

## Design Decisions

### Concurrency

I chose to use a fan-in/fan-out model using Go channels. An alternative would be running workers as separate processes
which require serialization/deserialization of messages being passed in between processes, which I feel is an overkill
for this particular exercise.

### Core abstractions

There are three types of main abstractions: `TimescaleBench`, `WorkerPool` and `Worker`.
- `TimescaleBench`  the main harness of the tool, responsible for:
  - parse the input csv file
  - for each row in the csv file, create a job with the specified `QueryParam`
  - send the job to the worker pool
  - collect the result from the worker pool and print out the formatted aggregation result
- `WorkerPool`  manager of workers, responsible for:
  - determine which job is dispatched to which worker
  - actually dispatch the job to the worker (fan-out)
  - collect query result from all workers (fan-in)
- `Worker` the worker goroutine that's responsible for:
  - connect to the database and execute the queries
  - time the query execution
  - post the query result to the results channel to be collected by the worker pool

### Query Routing

The requirement says the queries for the same host should go to the same worker.

I opted for a simpler hash bucket scheme where the hostname of the query is hashed, and the worker is selected based on
the bucket the hashed integer falls into (hash mod #workers).

I didn't go with consistent hashing because:
- it's more complicated to implement
- the benefit of consistent hashing is when you have workers joining/leaving frequently, in which case consistent hashing
minimizes the number of items need to be rehashed, but this is not a requirement for this assignment.
