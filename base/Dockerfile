#FROM alpine:3.18
#WORKDIR /app
#RUN wget https://dl.google.com/go/go1.19.7.linux-amd64.tar.gz
#RUN tar -C /usr/local -xzf go1.19.7.linux-amd64.tar.gz
#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
#ENV PATH="/usr/local/go/bin:${PATH}"
#RUN go version
FROM golang:1.21.7 AS builder
#WORKDIR /go/src/app
#
#RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources
#
#
#RUN apt-get update && apt-get install -y make nodejs npm && \
#    npm install -g yarn
#
WORKDIR /go/src/app

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    GOPRIVATE=github.com/SplashtopInc \
    GOOS=linux \
    GOARCH=amd64
#
#COPY ../../../../../ssw-3rdparty-fleet /fleet
#RUN cd /fleet && make deps-js generate

RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN dlv --help
ENTRYPOINT ["/bin/sh" ,"-c", "sleep 1000"]