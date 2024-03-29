FROM golang:1.21.0-bullseye AS builder

ARG App_Version

RUN apt-get update && apt-get --no-install-recommends -y install musl musl-dev musl-tools

WORKDIR /build

COPY . .

RUN go env -w CGO_ENABLED=1 GOOS=linux CC=/usr/bin/musl-gcc
RUN go get -d -v
RUN go build -v -ldflags="-linkmode external -extldflags=-static -w -s -X 'easy-image-cdn.pcpl2lab.ovh/app/build.Version=${App_Version}' -X 'easy-image-cdn.pcpl2lab.ovh/app/build.Time=$(date)'" -o imageCdn .
RUN go test -v ./imageConverter

RUN mkdir -p images
RUN touch images/dontRemoveMe.txt
RUN mkdir -p logs
RUN touch logs/dontRemoveMe.txt

FROM busybox:1.36.1 AS builder-user

RUN addgroup -g 10002 appUser && \
    adduser -D -u 10003 -G appUser appUser

FROM gcr.io/distroless/static-debian11
COPY --from=builder --chown=10003:10002 /build/imageCdn /
COPY --from=builder-user /etc/passwd /etc/passwd
COPY --from=builder --chown=10003:10002 /build/logs /var/log/eic/
COPY --from=builder --chown=10003:10002 /build/images /var/lib/images/

ENV IN_DOCKER=1 \
    API_KEY="00000000-0000-0000-0000-000000000000" \
    API_KEY_HEADER="key" \
    CONVERT_TO_RES="1024x720,800x600" \
    MAX_FILE_SIZE=10 \
    CACHE_TIME=30 \
    EXPVAR_ENABLED=0 \
    PPROF_ENABLED=0

EXPOSE 9324
EXPOSE 9555
EXPOSE 9125

USER appUser
ENTRYPOINT ["/imageCdn"]
