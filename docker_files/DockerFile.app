# docker_files/DockerFile.app
FROM golang:1.25.5

WORKDIR /usr/src/gomall

ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.google.cn

COPY app/app/go.mod app/app/go.sum ./app/app/
COPY rpc_gen ./rpc_gen
COPY common ./common

RUN cd app/app && go mod download && go mod verify

COPY app/app ./app/app

RUN cd app/app && go build -v -o server

WORKDIR /usr/src/gomall/app/app
EXPOSE 8080
CMD ["./server"]
