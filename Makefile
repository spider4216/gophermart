APP_NAME=gophermart

DATABASE_URI?=postgres://postgres:postgres@localhost:5432/gophermart

run:
	DATABASE_URI=${DATABASE_URI} go run ./cmd/gophermart -l debug

run-acc:
	./cmd/accrual/accrual_linux_amd64

test:
	go test ./...

test-v:
	go test ./... -v

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

build:
	go build -o ./cmd/gophermart ./cmd/gophermart

migration-gen:
	migrate create -ext sql -dir ./migrations -seq $(name)

migrate-up:
	migrate -path ./migrations -database $(DATABASE_URI) up $(ver)

migrate-down:
	migrate -path ./migrations -database $(DATABASE_URI) down $(ver)

migrate-force:
	migrate -path ./migrations -database $(DATABASE_URI) force $(ver)
