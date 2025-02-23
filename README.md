# Service description
Example service is existing to build from box ready to production service.
The service has the ability to provide data both via REST API and gRPC simultaneously.

## Table of Contents
- [Development](#development)
  - [Project structure](#project-structure)
  - [How to run locally and debug](#how-to-run-locally-and-debug)
  - [How to create and execute DB migrations](#how-to-create-and-execute-db-migrations)
  - [Deploy description](#deploy-description)
- [Monitoring and tracing](#monitoring)
- [Questions/Feedback](#questions-or-feedback)

# Development
## Project structure
- [docker](build%2Fapp%2Fdocker) - contains Dockerfile to build the app.
- [migrations](build%2Fapp%2Fmigrations) - contains migrations for DB.
- [cmd](cmd) - contains main function.
- [gen](gen) - contains generated code from .proto files and server implementation, swagger docs and server implementation.
- [k8s](k8s) - contains `helm` folder with `values.yaml` and `custom_manifests` folder which you can populate later, and they will be auto applied via `kubectl apply` in the pipeline
- [pkg](pkg) - contains external clients and generated clients with there protos.
- [src](src) - contains source code of a sample app.
- [.golangci.yml](.golangci.yml) - describes linter policy.
- [.quazar.yml](.quazar.yml) - describes Quazar integration.
- [makefile](makefile) - contains all needed scripts for local debug and coding.

## How to run locally and debug
You can debug an application locally.
To do this, you should:
- Go to the [makefile](makefile)
- Run `local-deps-up` script. All needed dependencies will start.
- Run `local-migrations-up` script. All needed migrations will be processed.
- Copy all [env vars](build/local/.env) from `./build/local/.env`.
- Run application with environments you copied.
- Profit. Now you can make queries to the service with URL `localhost:8000`

## How to create and execute DB migrations
We use migration tool to create changes in DB. To create a new migration:
- Go to [makefile](makefile)
- Run `local-create-migration` with specified name
- Go to [migrations](build/app/migrations)
- Fill created files with needed scripts
- Debug this the script locally with `local-deps-up` and `local-deps-down` scripts
- Migrations will be applied automatically on RC after a main pipeline succeeded
- Migrations will be applied automatically on PROD after a tag pipeline succeeded

## Monitoring and tracing
The Service contains predefined metrics which can be pulled from /metrics path, and it contains predefined tracing with OpenTelemetry.

## Questions or feedback?
For service question contact ingvar@mattis.dev.
