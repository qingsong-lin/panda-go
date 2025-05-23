FROM bdqi/go_builder:latest as builder
#FROM  alpine-3.18:v1 as builder
#
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on

WORKDIR /app
#ARG DIR
#RUN echo ${DIR}
COPY . /app
#RUN go env
RUN ls
RUN pwd
#RUN go mod init panda-go
#RUN go mod tidy
WORKDIR /app/panda-export
RUN pwd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -gcflags="all=-N -l" . #-gcflags="all=-N -l"
#FROM  alpine-3.18:v1
RUN ls /app
RUN ls /app/panda-export


FROM alpine:3.18
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
#RUN apk update
#RUN apk add curl
EXPOSE 8080
EXPOSE 1234
EXPOSE 50051
#RUN sed -i 's/https/http/' /etc/apk/repositories
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
#RUN  apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone # buildkit
#RUN  apk add --no-cache ca-certificates
#RUN  apk add --update sudo # buildkit

#ENV TZ=Asia/Shanghai
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
#    && apk update \
#    && apk add --no-cache tzdata gcc g++ libffi-dev musl-dev openssl-dev make linux-headers libc-dev libc6-compat binutils \
#    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime \
#    && echo $TZ > /etc/timezone
#RUN apk --no-cache add ca-certificates wget
#RUN wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub
#RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.28-r0/glibc-2.28-r0.apk
#RUN apk add glibc-2.28-r0.apk
WORKDIR /app
COPY --from=builder /app/panda-export/panda-export /app/
CMD ["/app/panda-export"]