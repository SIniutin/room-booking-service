DOCKER_HOST ?= $(shell docker context inspect --format '{{(index .Endpoints "docker").Host}}' 2>/dev/null)
TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE ?= /var/run/docker.sock

export DOCKER_HOST
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE

up:
	docker compose up --build -d

keys:
	./scripts/generate-dev-keys.sh ./secrets

rotate-keys:
	FORCE_REGENERATE=1 docker compose run --rm keys

down:
	docker compose down

logs:
	docker compose logs -f api

seed:
	@echo "Seeding test data..."
	@curl -s -X POST http://localhost:8080/dummyLogin -H 'Content-Type: application/json' \
		-d '{"role":"admin"}' | jq .
	@curl -s -X POST http://localhost:8080/dummyLogin -H 'Content-Type: application/json' \
		-d '{"role":"user"}' | jq .

reset-db:
	docker compose up -d postgres
	docker compose exec postgres psql -U postgres -d booking -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

unit:
	@packages="$$(go list ./internal/... ./cmd/api/... | grep -v '/internal/repository/pg$$')"; \
	go test -race -count=1 $$packages

integration:
	go test -count=1 -timeout=5m ./internal/repository/pg

test: unit integration

vet:
	go vet ./...

e2e:
	go test -tags=e2e -v -timeout=120s ./e2e/...

lint:
	golangci-lint run ./...

load-slots:
	k6 run load/k6_slots_list.js

.PHONY: up keys rotate-keys down logs seed reset-db unit integration test vet e2e lint load-slots
