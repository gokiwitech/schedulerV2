# builder image
FROM golang:1.22-alpine as builder
RUN mkdir /build
WORKDIR /build
COPY . .
COPY .env /app/.env
RUN GOOS=linux GOARCH=amd64 go build -v -o schedulerV2 .


# executable
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /build/schedulerV2 .
COPY .env /app/.env
EXPOSE 9999
CMD ["./schedulerV2", "-port", ":9999"]
