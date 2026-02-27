FROM golang:1.26-alpine AS builder
ARG SERVICE
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /out/service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
