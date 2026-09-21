.DEFAULT_GOAL := help

## up: build and start the api and database
up:
	docker compose up --build

## down: stop everything and remove the data
down:
	docker compose down -v

## run: run the api locally (start the database first with `docker compose up db`)
run:
	go run ./cmd/api

## test: run the unit tests
test:
	go test -race ./...

## newman: run the postman collection against a running api
newman:
	docker run --rm --add-host=host.docker.internal:host-gateway -v "$(CURDIR)/tests:/etc/newman" \
		postman/newman:alpine run hotel-booking.postman_collection.json \
		--env-var base_url=http://host.docker.internal:8080

## help: show this help
help:
	@echo "Usage: make <target>"
	@echo
	@grep -E '^## [a-z]+:' $(MAKEFILE_LIST) | sed 's/^## //' | awk -F': ' '{printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}'

.PHONY: up down run test newman help
