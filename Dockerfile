# syntax=docker/dockerfile:experimental
FROM golangci/golangci-lint:v1.57.0 AS base

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go mod tidy

FROM base as lint
RUN golangci-lint run --timeout 10m0s ./...

FROM base as test
RUN go test -v -coverprofile=cover.out ./...
RUN go tool cover -func=cover.out

FROM base as builder
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:3.19
RUN adduser -D nonroot
WORKDIR /app
RUN chown nonroot:nonroot /app
USER nonroot
RUN touch .env
COPY --chown=nonroot:nonroot --from=builder /app/main .
COPY --chown=nonroot:nonroot --from=builder /app/docs/* ./docs/
CMD ["./main"]
