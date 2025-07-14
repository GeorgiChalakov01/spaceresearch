BINARY := app
.PHONY: clear compile run

clear:
	@clear

compile:
	@echo "Generating templ files..."
	go tool templ generate ./...
	@echo "Building binary..."
	CGO_ENABLED=0 go build -o $(BINARY) .

run: 
	@echo "Starting..."
	go run main.go
