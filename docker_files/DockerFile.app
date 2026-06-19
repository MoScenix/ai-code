# docker_files/DockerFile.app
FROM golang:1.25.5

WORKDIR /usr/src/gomall

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
ENV GOTOOLCHAIN=local
ENV GOWORK=off

COPY app/app/go.mod app/app/go.sum ./app/app/
COPY rpc_gen ./rpc_gen
COPY common ./common

RUN cd app/app && go mod download && go mod verify

COPY app/app ./app/app

RUN cd app/app && go build -mod=readonly -v -o server

WORKDIR /usr/src/gomall/app/app
EXPOSE 8888
CMD ["./server"]
