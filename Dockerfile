# builder image
#FROM golang:1.22-alpine as builder
#RUN mkdir /build
#WORKDIR /build
#COPY . . 
#RUN GOOS=linux GOARCH=amd64 go build -v -o schedulerV2 .

# executable
#FROM alpine:latest
#WORKDIR /root/
#COPY --from=builder /build/schedulerV2 .
#COPY .env /root/.env
#EXPOSE 9999
#CMD ["./schedulerV2", "-port", ":9999"]


# Use an appropriate base image
FROM golang:1.22-alpine as builder

# Set the working directory
WORKDIR /build

# Copy all files into the container
COPY . .

# Build the application
RUN GOOS=linux GOARCH=amd64 go build -v -o schedulerV2 .

# Executable
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /build/schedulerV2 .

# Copy the config file passed as a build argument
ARG CONFIG_FILE
COPY config/environments/${CONFIG_FILE} ./config/environments/${CONFIG_FILE}

# Expose the port
EXPOSE 9999

# Command to run the application
CMD ["./schedulerV2", "-port", ":9999"]
