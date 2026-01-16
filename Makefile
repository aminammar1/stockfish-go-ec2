APP_NAME=stockfish-ec2-service
CMD_DIR=./cmd/server
CLI_DIR=./cmd/cli

.PHONY: tidy build run swag cli-build cli-run

tidy:
	go mod tidy

build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)

cli-build:
	go build -o bin/$(APP_NAME)-cli $(CLI_DIR)

run:
	go run $(CMD_DIR)

cli-run:
	go run $(CLI_DIR)

swag:
	swag init -g cmd/server/main.go -o docs

help:
	@echo "Makefile commands:"
	@echo "  tidy   - Clean up go.mod and go.sum files"
	@echo "  build  - Build the application binary"
	@echo "  run    - Run the application"
	@echo "  swag   - Generate Swagger documentation"
