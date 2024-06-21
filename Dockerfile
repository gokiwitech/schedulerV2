# builder image
FROM golang:1.19-alpine as builder
RUN mkdir /build
WORKDIR /build
COPY . /build/
RUN CGO_ENABLED=0 go build -a -o schedulerV2 .


# executable
EXPOSE 9999
# Run the executable
CMD ["./schedulerV2", "-port", ":9999"]
