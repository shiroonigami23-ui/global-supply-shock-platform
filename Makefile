SHELL := /bin/sh

.PHONY: tidy fmt test run-ingest run-risk run-alert run-api up down build k8s-apply k8s-delete

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

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

k8s-apply:
	kubectl apply -k deploy/k8s

k8s-delete:
	kubectl delete -k deploy/k8s
