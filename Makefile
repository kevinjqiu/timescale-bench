data:
	curl -sLo data.tar.gz "https://www.dropbox.com/s/17mr38w21yhgjjl/TimescaleDB_coding_assignment-RD_eng_setup.tar.gz?dl=1"

start-timescale:
	docker run -d --name timescaledb -p 5432:5432 -e POSTGRES_PASSWORD=password timescale/timescaledb:2.0.0-pg12
