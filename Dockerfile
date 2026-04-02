FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /yodahunters ./cmd/backend

FROM alpine:latest
WORKDIR /app
COPY --from=builder /yodahunters /app/yodahunters
COPY static /app/static
COPY templates /app/templates
EXPOSE 8080
ENTRYPOINT ["/app/yodahunters"]
