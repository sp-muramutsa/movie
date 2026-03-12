FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o main ./movie/cmd/*.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder app/main ./
COPY movie/configs/default.yaml ../configs/default.yaml
RUN chmod +x ./main
EXPOSE 8083
CMD ["./main"]