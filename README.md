# Service description
Example Service is designed for quickly creating production-ready microservices.
It natively supports two APIs out of the box: gRPC and REST, both generated from .proto files.

## Table of Contents
- [Development](#development)
  - [Project structure](#project-structure)
  - [How to run locally and debug](#running-locally-and-debugging)
  - [How to create and execute DB migrations](#creating-and-executing-database-migrations)
- [Monitoring and tracing](#monitoring-and-tracing)
- [Questions/Feedback](#questions-and-feedback)

# Development
## Project structure
- [docker](build/app/docker) - Contains the Dockerfile for building the application.
- [migrations](build/app/migrations) - Contains database migration scripts.
- [cmd](cmd) - Contains the application's entry point.
- [gen](gen) - Stores generated code from .proto files, including server implementations and Swagger documentation.
- [src](src) - Contains the application's source code.
- [.golangci.yml](.golangci.yml)(.golangci.yml) - Defines linting rules and policies.
- [makefile](makefile) - Contains essential scripts for local development and debugging.

## Running Locally and Debugging
To run and debug the application locally, follow these steps:
- Open the [Makefile](makefile).
- Run the `local-deps-up` script to start all required dependencies.
- Run the `local-migrations-up` script to apply necessary database migrations.
- Copy all environment variables from [env vars](build/local/.env).
- Start the application using the copied environment variables.
- Done! You can now make requests to the service at `http://localhost:8000`.

## Creating and Executing Database Migrations
This service uses a migration tool for database schema changes.
To create a new migration, follow these steps:
- Open the [makefile](makefile).
- Run the `local-create-migration` script with a specified name.
- Navigate to the [migrations](build/app/migrations) directory.
- Edit the newly created migration files and add the required SQL scripts.
- (Optional) Start a local database instance using local-deps-up and stop it using local-deps-down.
- Apply and test the migrations locally using `local-migrations-up` and `local-migrations-down`.

## Monitoring and tracing
The service exposes predefined metrics at the /metrics endpoint and includes tracing via OpenTelemetry.

## Questions and feedback?
For any questions regarding this service, contact ingvar@mattis.dev.
