FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o main ./rating/cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main ./
COPY rating/configs/default.yaml ../configs/default.yaml
RUN chmod +x ./main
EXPOSE 8082
CMD ["./main"]
