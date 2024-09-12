FROM golang:1.23.1-alpine3.20

RUN apk add --update --no-cache git zsh build-base openssh
RUN sh -c "$(wget -nv -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.2.0/zsh-in-docker.sh)" -- \
    -t robbyrussell \
    -p git \
    -p copyfile \
    -p extract \
    -p colorize \
    -p golang

RUN go env -w GO111MODULE=auto
RUN go env -w CGO_ENABLED=1
ENV CGO_ENABLED=1
RUN go install github.com/go-delve/delve/cmd/dlv@v1.23.0
#RUN go install github.com/ramya-rao-a/go-outline@v1.0.0 
