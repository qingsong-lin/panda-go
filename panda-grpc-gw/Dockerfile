FROM bdqi/go_builder:latest as builder
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on

WORKDIR /tmp
RUN wget http://192.168.50.245:8000/protoc-25.1-linux-x86_64.zip
RUN unzip protoc-25.1-linux-x86_64.zip
RUN cp bin/protoc $GOPATH/bin

#ARG DIR
#RUN echo ${DIR}
WORKDIR /app
COPY . /app
WORKDIR /app/proto
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
RUN go get -u github.com/golang/protobuf/protoc-gen-go
RUN go install github.com/golang/protobuf/protoc-gen-go
RUN go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
#RUN go env
RUN VERSION=$(ls $(go env GOMODCACHE)/github.com/grpc-ecosystem/grpc-gateway/ | head -n 1) && protoc -I=/tmp/include -I. -I$(go env GOMODCACHE)/github.com/grpc-ecosystem/grpc-gateway/$VERSION --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:. --grpc-gateway_out=logtostderr=true:. --openapiv2_out=logtostderr=true:.  *.proto
#RUN ls
#RUN pwd
##RUN go mod init panda-go
##RUN go mod tidy
WORKDIR /app/panda-grpc-gw
#RUN pwd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -gcflags="all=-N -l" . #-gcflags="all=-N -l"
##FROM  alpine-3.18:v1
#RUN ls /app
#RUN ls /app/panda-auth
FROM alpine:3.18


EXPOSE 8080
EXPOSE 50051
##RUN sed -i 's/https/http/' /etc/apk/repositories
##RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
##RUN  apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone # buildkit
##RUN  apk add --no-cache ca-certificates
##RUN  apk add --update sudo # buildkit
#
##ENV TZ=Asia/Shanghai
##RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
##    && apk update \
##    && apk add --no-cache tzdata gcc g++ libffi-dev musl-dev openssl-dev make linux-headers libc-dev libc6-compat binutils \
##    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime \
##    && echo $TZ > /etc/timezone
##RUN apk --no-cache add ca-certificates wget
##RUN wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub
##RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.28-r0/glibc-2.28-r0.apk
##RUN apk add glibc-2.28-r0.apk
#WORKDIR /app
COPY --from=builder /app/panda-grpc-gw/panda-grpc-gw /app/
COPY --from=builder /app/proto /app/proto
WORKDIR /app
CMD ["/app/panda-grpc-gw"]