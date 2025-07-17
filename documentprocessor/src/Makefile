BINARY := app
.PHONY: clear compile run

clear:
	@clear

compile:
	@echo "Building binary..."
	CGO_ENABLED=1 GOOS=linux go build -ldflags="-extldflags=-static" -o $(BINARY) .

run: 
	@echo "Starting..."
	go run main.go
