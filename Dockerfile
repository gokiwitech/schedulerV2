# builder image
FROM golang:1.22-alpine as builder
RUN mkdir /build
WORKDIR /build
COPY . .  # Copy all files including .env to /build directory
RUN GOOS=linux GOARCH=amd64 go build -v -o schedulerV2 .

# executable
FROM alpine:latest
RUN mkdir /root
WORKDIR /root
COPY --from=builder /build/schedulerV2 .
COPY .env /root/.env  # Ensure the .env file is copied to the final image
EXPOSE 9999
CMD ["./schedulerV2", "-port", ":9999"]
