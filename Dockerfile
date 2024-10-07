# builder image
FROM golang:1.22-alpine as builder
RUN mkdir /build
WORKDIR /build
COPY . .
COPY .env /build/.env  # Ensure .env is copied to the build folder
RUN GOOS=linux GOARCH=amd64 go build -v -o schedulerV2 .

# executable
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /build/schedulerV2 .
COPY --from=builder /build/.env .  # Copy .env to the working directory
EXPOSE 9999
CMD ["./schedulerV2", "-port", ":9999"]
