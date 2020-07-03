FROM golang:1.13-alpine as builder
ARG VERSION=0.0.1
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
#ENV GOARCH=arm

# build
WORKDIR /go/src/sfc-controller/main
COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on go mod download
COPY . .
RUN go install -ldflags "-s -w -X main.version=$VERSION" sfc-controller

# runtime image
FROM gcr.io/google_containers/ubuntu-slim:0.14
COPY --from=builder /go/bin/sfc-controller /usr/bin/sfc-controller
EXPOSE 8100
ENTRYPOINT ["sfc-controller"]

# Usage:

#docker build -t jpedro1992/sfc-controller:0.0.4 .
#docker push jpedro1992/sfc-controller:0.0.4

