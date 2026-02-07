# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/zikir-hatmi-backend ./

# Runtime stage
FROM alpine:3.20

RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /bin/zikir-hatmi-backend /app/server

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/server"]
