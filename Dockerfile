FROM golang:1.25-alpine AS builder

# No gcc/musl-dev needed: the sqlite driver is pure Go (modernc.org/sqlite),
# so the build is CGO-free, fast, and low-memory.

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produces a fully static binary and keeps the build light enough
# that default parallelism is safe (no more -p 1 serialisation).
RUN CGO_ENABLED=0 GOOS=linux go build -o cloud-guard ./cmd/server

FROM alpine:latest

WORKDIR /root/

# CA certificates for AWS HTTPS. No sqlite-libs needed - driver is pure Go.
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/cloud-guard .
COPY --from=builder /app/web ./web
# Needed by GET /cloudformation.yaml - without this the onboarding download 404s.
COPY --from=builder /app/deployments ./deployments

EXPOSE 8080

CMD ["./cloud-guard"]
