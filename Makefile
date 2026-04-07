up:
	docker compose up --build -d

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

test:
	go test ./...

e2e:
	go test -tags=e2e -v -timeout=120s ./e2e/...

lint:
	golangci-lint run ./...

load-slots:
	k6 run load/k6_slots_list.js

.PHONY: up down logs seed reset-db test e2e lint load-slots
