.PHONY: build run test clean docker dev

build:
	go build -o bin/proxi-pulse ./cmd/proxi-pulse

run: build
	./bin/proxi-pulse

test:
	go test ./... -v -count=1

clean:
	rm -rf bin/

docker:
	docker build -t proxi-pulse:latest .

dev:
	air
