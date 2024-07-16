binary=./bin/api

build:
	@go build -o $(binary)

run: build
	@$(binary)