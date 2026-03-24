FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY metadata ./metadata
COPY pkg ./pkg
COPY api ./api
COPY gen ./gen
COPY internal ./internal
COPY rating ./rating
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV TALLY_PROMETHEUS_SANITIZE_NAME=true
RUN go build -trimpath -ldflags="-s -w" -o main ./metadata/cmd

FROM alpine:latest
WORKDIR /app
ENV TALLY_PROMETHEUS_SANITIZE_NAME=true
COPY --from=builder /app/main ./
COPY metadata/configs/default.yaml ../configs/default.yaml
COPY metadata/configs/server.cert ../configs/server.cert
COPY metadata/configs/server.key ../configs/server.key
RUN chmod +x ./main
EXPOSE 8081
CMD ["./main"]
