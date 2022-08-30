FROM golang:1.18.5-alpine3.16 AS builder

WORKDIR /build

COPY . .

RUN apk update && \
    apk add git build-base

RUN go env -w CGO_ENABLED=1 GOOS=linux
RUN go get -d -v
RUN go build -v -ldflags="-w -s" -o imageCdn .
RUN go test -v ./imageConverter

RUN mkdir -p images
RUN touch images/dontRemoveMe.txt

FROM busybox AS builder-user

RUN addgroup -g 10002 appUser && \
    adduser -D -u 10003 -G appUser appUser

FROM alpine:3.16
COPY --from=builder /build/imageCdn /
COPY --from=builder-user /etc/passwd /etc/passwd
COPY --from=builder --chown=10003:10002 /build/images /var/lib/images/

ENV IN_DOCKER=1 \
    API_KEY="00000000-0000-0000-0000-000000000000" \
    API_KEY_HEADER="key" \
    CONVERT_TO_RES="1024x720,800x600" \
    MAX_FILE_SIZE=10 \
    CACHE_TIME=30

EXPOSE 9324
EXPOSE 9555

USER appUser
ENTRYPOINT ["/imageCdn"]
