# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go/go.mod go/go.sum ./
RUN go mod download

COPY go/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/m01-state-engine ./cmd/server/

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/m01-state-engine /usr/local/bin/

EXPOSE 50051

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:50051/health || exit 1

CMD ["m01-state-engine"]
