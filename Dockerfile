FROM cache-registry.caas.intel.com/cache/library/golang:1.20
ENV http_proxy=http://proxy-chain.intel.com:912
ENV https_proxy=http://proxy-chain.intel.com:912
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
ENV TZ=Asia/Taipei
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apt-get update && apt-get install -y ffmpeg
RUN go mod download
