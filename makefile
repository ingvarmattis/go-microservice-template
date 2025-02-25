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
## download-dependencies is needed to download ./gen/protos/google/api/* files
## you need to manually move google/api/annotations.proto and google/api/http.proto to gen/protos
download-dependencies:
	git clone https://github.com/googleapis/googleapis.git

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
      ./gen/protos/*/*.proto

#Unit tests
unit-tests:
	go test -v -tags unit_tests ./... ./...

#Build
build-application:
	docker build -t example -f ./build/app/docker/Dockerfile .
