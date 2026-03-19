# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . ./

# Build
WORKDIR /src/cmd/server
RUN go build -o /out/halleyx-server

# Runtime stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /out/halleyx-server .
COPY --from=builder /src/frontend ./frontend
COPY --from=builder /src/migrations ./migrations

EXPOSE 8080

CMD ["./halleyx-server"]
