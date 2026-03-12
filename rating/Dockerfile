FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o main ./rating/cmd/*.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/rating/configs/default.yaml /configs/default.yaml
EXPOSE 8082
CMD ["./main"]