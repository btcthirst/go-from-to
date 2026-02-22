.PHONY: clean test build run

clean:
	rm -rf build dist *.egg-info

test:
	go test -v ./...

build:
	go build -o bin/app cmd/main.go

run: build
	./bin/app