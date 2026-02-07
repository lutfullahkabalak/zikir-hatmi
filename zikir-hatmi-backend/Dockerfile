# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN set -eux; \
	goarm=""; \
	if [ "${TARGETARCH}" = "arm" ]; then \
		goarm="${TARGETVARIANT#v}"; \
	fi; \
	CGO_ENABLED=0 \
	GOOS="${TARGETOS}" \
	GOARCH="${TARGETARCH}" \
	${goarm:+GOARM=${goarm}} \
	go build -trimpath -ldflags="-s -w" -o /bin/zikir-hatmi-backend ./

# Runtime stage
FROM --platform=$TARGETPLATFORM alpine:3.20

RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /bin/zikir-hatmi-backend /app/server

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/server"]
