golangci-lint-run:
	golangci-lint run -c .golangci.yml

-include .env
MIGRATIONS_DIR=./internal/server/infra/db/migrations

migrate:
	migrate -path "$(MIGRATIONS_DIR)" -database $(DSN) $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

migration:
	migrate create -ext sql -seq -digits 3 -dir "$(MIGRATIONS_DIR)" $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

coverprofile-services:
	go test ./... -covermode=count -coverprofile cover.out.tmp && cat cover.out.tmp \
		| grep -v -e "test" -e "mocks" -e "pkg/proto" -e "pkg/logging" -e "/cmd" -e "internal/client/tui" \
		 -e "internal/client/app" -e "internal/client/client" -e "internal/server/server" -e "config" > cover.out \
 		&& rm cover.out.tmp && go tool cover -html cover.out -o coverprofile.html && go tool cover -func cover.out

proto-auth:
	protoc --go_out=. --go-grpc_out=. api/proto/auth.proto

proto-secret:
	protoc --go_out=. --go-grpc_out=. api/proto/secret.proto

proto: proto-auth
proto: proto-secret

.PHONY: certs
certs:
	mkdir -p certs
	go run ./cmd/gencerts/main.go

ifneq ($(filter $(MAKECMDGOALS),migrate migration),)
%:
	@true
endif
