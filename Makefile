.PHONY: run build clean test test-verbose test-short test-coverage benchmark

build:
	go build -o bin/app cmd/app/main.go

run: build
	./bin/app

clean:
	rm -f bin/app summery.xlsx

# Запуск всіх тестів
test:
	go test ./...

# Запуск тестів з детальним виводом
test-verbose:
	go test -v ./...

# Запуск швидких тестів (без інтеграційних)
test-short:
	go test -short ./...

# Запуск тестів із мірами покриття
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Запуск бенчмарків
benchmark:
	go test -bench=. -benchmem ./...

# Запуск бенчмарків для конкретного пакету
benchmark-reader:
	go test -bench=. -benchmem ./internal/excel/reader

benchmark-config:
	go test -bench=. -benchmem ./internal/config

benchmark-app:
	go test -bench=. -benchmem ./internal/app
