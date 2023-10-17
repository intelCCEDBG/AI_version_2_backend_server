FROM golang:1.20
ENV http_proxy=http://proxy-chain.intel.com:911
ENV https_proxy=http://proxy-chain.intel.com:911
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .

RUN go mod download
