.PHONY: run build clean

build:
	go build -o bin/app cmd/app/main.go

run: build
	./bin/app

clean:
	rm -f bin/app summery.xlsx
