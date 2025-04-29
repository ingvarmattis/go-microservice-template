#!/bin/bash

# === Настройки ===
OLD_SERVICE_NAME="example-service"
NEW_SERVICE_NAME="go-auth"

OLD_SCHEMA_NAME="example"
NEW_SCHEMA_NAME="auth"

USE_POSTGRES_DB=true  # false if you will not user a postgres db

# === Step 1. Dockerfile ===
sed -i "s|$OLD_SERVICE_NAME|$NEW_SERVICE_NAME|g" ./build/app/docker/Dockerfile

# === Step 2. Postgres ===
if [ "$USE_POSTGRES_DB" = true ]; then
    sed -i "s|$OLD_SCHEMA_NAME|$NEW_SCHEMA_NAME|g" ./build/app/migrations/*.up.sql
    sed -i "s|$OLD_SCHEMA_NAME|$NEW_SCHEMA_NAME|g" ./build/app/migrations/*.down.sql
else
    rm -rf ./build/app/migrations
    sed -i '/^postgres:/,/^[^[:space:]]/d' ./build/local/docker-compose.yaml
    sed -i '/#PostgresConfig/,+1d' ./build/local/.env
    sed -i '/^'"${OLD_SERVICE_NAME^^}"'_POSTGRES_URL=/d' ./build/local/.env
fi

# === Step 3. .env file ===
# Rename variable values
sed -i "s|${OLD_SERVICE_NAME^^}_NAME=.*|${NEW_SERVICE_NAME^^}_NAME=${NEW_SERVICE_NAME}|g" ./build/local/.env
# Rename variable prefixes
sed -i "s|${OLD_SERVICE_NAME^^}|${NEW_SERVICE_NAME^^}|g" ./build/local/.env

# === Step 4. config.go ===
sed -i "s|${OLD_SERVICE_NAME^^}|${NEW_SERVICE_NAME^^}|g" ./src/config/config.go

# === Step 5. docker-compose.yaml ===
sed -i "s|name: ['\"]\?$OLD_SERVICE_NAME['\"]\?|name: '$NEW_SERVICE_NAME'|g" ./build/local/docker-compose.yaml

# === Step 6. Rename directory rpctransport ===
mv ./src/rpctransport/${OLD_SCHEMA_NAME} ./src/rpctransport/${NEW_SCHEMA_NAME}

# === Step 7. Rename directory services ===
mv ./src/services/${OLD_SCHEMA_NAME} ./src/services/${NEW_SCHEMA_NAME}

echo "Done."
