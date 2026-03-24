FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY rating ./rating
COPY pkg ./pkg
COPY api ./api
COPY gen ./gen
COPY internal ./internal
COPY metadata ./metadata
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV TALLY_PROMETHEUS_SANITIZE_NAME=true
RUN go build -trimpath -ldflags="-s -w" -o main ./rating/cmd

FROM alpine:latest
WORKDIR /app
ENV TALLY_PROMETHEUS_SANITIZE_NAME=true
COPY --from=builder /app/main ./
COPY rating/configs/default.yaml ../configs/default.yaml
RUN chmod +x ./main
EXPOSE 8082
CMD ["./main"]
