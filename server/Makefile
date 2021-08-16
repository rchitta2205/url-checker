.PHONY: dependency unit-test integration-test up down clear

dependency:
	@go get -v ./...

unit-test: dependency
	@go test -v -short ./...

up:
	@docker-compose up -d

down:
	@docker-compose down

integration-test: up dependency
	@go test -v ./...

clear: down
