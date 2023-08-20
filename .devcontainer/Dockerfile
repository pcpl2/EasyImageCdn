FROM golang:alpine3.18

RUN apk add --update --no-cache git zsh build-base openssh
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.1.2/zsh-in-docker.sh)" -- \
    -t robbyrussell \
    -p git \
    -p copyfile \
    -p extract \
    -p colorize \
    -p golang

RUN go env -w GO111MODULE=auto
RUN go env -w CGO_ENABLED=1
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/ramya-rao-a/go-outline@v0.0.0-20210608161538-9736a4bde949 