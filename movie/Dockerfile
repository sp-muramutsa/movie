FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./movie/cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main ./
COPY movie/configs/default.yaml ../configs/default.yaml
COPY movie/configs/server.cert ../configs/server.cert
COPY movie/configs/server.key ../configs/server.key
RUN chmod +x ./main
EXPOSE 8083
CMD ["./main"]
