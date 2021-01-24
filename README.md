# timescale-bench

Benchmark timescaledb `select` performance.

## Usage

    timescale-bench --input query_params.csv --workers 10

### Flags

* `--input` - path to the query params csv file. Use `-` to read the input from STDIN.
* `--workers` - number of concurrent workers

## Design Decisions

### Query Routing

The requirement says the queries for the same host should go to the same worker.

I opted for a simpler hash bucket scheme where the hostname of the query is hashed, and the worker is selected based on
the bucket the hashed integer falls into (hash mod #workers).

I didn't go with consistent hashing because:
- it's more complicated to implement
- the benefit of consistent hashing is when you have workers joining/leaving frequently, in which case consistent hashing
minimizes the number of items need to be rehashed, but this is not a requirement for this assignment.
  