FROM golang:alpine3.14

RUN apk add --update --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev git zsh build-base openssh
RUN go env -w GO111MODULE=off
RUN go get -u golang.org/x/tools/gopls 2>&1
RUN go get -u honnef.co/go/tools/... 2>&1
RUN go get -u golang.org/x/lint/golint 2>&1
RUN go get -u github.com/uudashr/gopkgs/v2/cmd/gopkgs 2>&1
RUN go get -u golang.org/x/tools/gopls 2>&1
RUN go get -u github.com/go-delve/delve/cmd/dlv 2>&1

RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.1.2/zsh-in-docker.sh)" -- \
    -t robbyrussell \
    -p git \
    -p copyfile \
    -p extract \
    -p colorize \
    -p golang

RUN go env -w GO111MODULE=auto
