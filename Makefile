.PHONY: build up down clear

build:
	@docker-compose build

up: build
	@docker-compose up

down:
	@docker-compose down

clear: down