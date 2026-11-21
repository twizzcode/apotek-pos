.PHONY: db-up db-down db-reset run test

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d

run:
	go run ./cmd/web

test:
	go test ./...
