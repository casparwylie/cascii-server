#!/bin/bash

echo "Clearing test database..."

mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p${DB_PASS} -e \
  "DROP DATABASE IF EXISTS ${DB_NAME}; CREATE DATABASE ${DB_NAME};"

echo "Preparing test database..."
migrate -source file://./migrations -database \
  "mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" up

echo "Starting server..."
./cascii-server &

sleep 3

echo "Running tests..."

go test
