#Prepare template
prepare-template:
	rm -r ./.git
	./scripts/replace_service.sh ./go-microservice-template old-service new-service

#Lint
go-lint:
	golangci-lint run

#Local dependencies
local-deps-up:
	docker compose -f ./build/local/docker-compose.yaml up -d

local-deps-down:
	docker compose -f ./build/local/docker-compose.yaml down

#Local migrations
local-create-migration:
	migrate create -ext sql -dir ./build/app/migrations #<migration_name>

local-migrations-up:
	migrate -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -path ./build/app/migrations up

local-migrations-down:
	yes | migrate -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -path ./build/app/migrations down

#Generate .proto
download-dependencies:
	if not exist gen\protos\google\api mkdir gen\protos\google\api
	curl -L -o gen\protos\google\api\annotations.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
	curl -L -o gen\protos\google\api\http.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto

generate-proto:
	protoc \
		--proto_path=./gen/protos \
		--proto_path=./gen/protos/google/api \
		--proto_path=./gen/protos/params \
		--go_out=. \
		--go-grpc_out=. \
		--grpc-gateway_out=. \
		--openapiv2_out=./gen/docs \
		--openapiv2_opt logtostderr=true \
		./gen/protos/*.proto \
		./gen/protos/params/*.proto

#Unit tests
export CGO_ENABLED=1
unit-tests:
	go test -v -tags unit_tests -race ./... ./...

benchmark-tests:
	go test -v -tags unit_tests -bench=. -benchmem ./... ./...

#Build
build-application:
	docker build -t example -f ./build/app/docker/Dockerfile .
