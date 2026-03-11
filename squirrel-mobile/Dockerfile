FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY server.go .
RUN go build -o server server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# 复制静态文件
COPY index.html .
COPY server .

EXPOSE 8080

CMD ["./server"]
