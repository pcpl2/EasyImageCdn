FROM golang:alpine3.14

RUN apk add --update --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev git zsh build-base

RUN go get -u -v golang.org/x/tools/gopls@latest 2>&1
RUN go get -u -v honnef.co/go/tools/... 2>&1
RUN go get -u -v golang.org/x/lint/golint 2>&1
RUN go get -u -v github.com/uudashr/gopkgs/v2/cmd/gopkgs 2>&1
RUN go get -u -v golang.org/x/tools/gopls@latest 2>&1
RUN go get -u -v github.com/go-delve/delve/cmd/dlv@master 2>&1

RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.1.2/zsh-in-docker.sh)" -- \
    -t robbyrussell \
    -p git \
    -p copyfile \
    -p extract \
    -p colorize \
    -p dotenv \
    -p golang
