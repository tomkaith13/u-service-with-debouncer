run:
	go run ./main.go
docker-build:
	docker build -t github.com/tomkaith13/u-service-with-debouncer .

clean:
	docker compose down
up:
	docker compose up
restart:
	make clean && make up
