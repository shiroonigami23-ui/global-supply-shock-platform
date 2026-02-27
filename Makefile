SHELL := /bin/sh

.PHONY: tidy run-ingest run-risk run-alert run-api up down build

tidy:
	go mod tidy

run-ingest:
	go run ./cmd/ingest

run-risk:
	go run ./cmd/risk-engine

run-alert:
	go run ./cmd/alert-service

run-api:
	go run ./cmd/query-api

up:
	docker compose up --build

down:
	docker compose down -v

build:
	go build ./...
