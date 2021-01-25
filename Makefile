build:
	go build -o bin/timescale-bench

data:
	curl -sLo data.tar.gz "https://www.dropbox.com/s/17mr38w21yhgjjl/TimescaleDB_coding_assignment-RD_eng_setup.tar.gz?dl=1"

start-timescale:
	docker run -d --name timescaledb -p 5432:5432 -e POSTGRES_PASSWORD=password timescale/timescaledb:2.0.0-pg12

shell:
	docker exec -it timescaledb bash

init-copy-file:
	docker cp data/ timescaledb:/tmp/

init-db:
	docker exec -it timescaledb sh -c 'psql -U postgres < /tmp/data/cpu_usage.sql'

init-data:
	docker exec -it timescaledb sh -c 'psql -U postgres -d homework -c "\COPY cpu_usage FROM /tmp/data/cpu_usage.csv CSV HEADER"'

init: init-copy-file init-db init-data

test:
	go test -v ./...
