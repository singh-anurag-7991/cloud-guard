FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO (sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary with CGO enabled.
# -p 1 limits the compiler to one package at a time. The aws-sdk-go-v2 ec2 package
# is huge (~900MB RSS to compile on its own); with the default parallelism two
# concurrent `compile` processes peak at ~1.5GB and OOM-kill small instances.
RUN CGO_ENABLED=1 GOOS=linux go build -p 1 -o cloud-guard ./cmd/server

FROM alpine:latest

WORKDIR /root/

# Install CA certificates (for AWS HTTPS) and sqlite libs
RUN apk --no-cache add ca-certificates sqlite-libs

COPY --from=builder /app/cloud-guard .
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./cloud-guard"]
