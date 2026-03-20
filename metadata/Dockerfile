FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./metadata/cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main ./
COPY metadata/configs/default.yaml ../configs/default.yaml
COPY metadata/configs/server.cert ../configs/server.cert
COPY metadata/configs/server.key ../configs/server.key
RUN chmod +x ./main
EXPOSE 8081
CMD ["./main"]
