FROM rust:1.86.0-alpine AS builder

ARG App_Version

RUN apk add --no-cache \
    build-base \
    musl-dev \
    openssl-dev \
    openssl-libs-static \
    pkgconf \
    git \
    libgcc \
    libstdc++
# Set `SYSROOT` to a dummy path (default is /usr) because pkg-config-rs *always*
# links those located in that path dynamically but we want static linking, c.f.
# https://github.com/rust-lang/pkg-config-rs/blob/54325785816695df031cef3b26b6a9a203bbc01b/src/lib.rs#L613
ENV SYSROOT=/dummy

WORKDIR /build

COPY . .

RUN cargo build --bins --release

RUN mkdir -p images
RUN touch images/dontRemoveMe.txt
RUN mkdir -p logs
RUN touch logs/dontRemoveMe.txt

FROM busybox:1.37.0 AS builder-user

RUN addgroup -g 10002 appUser && \
    adduser -D -u 10003 -G appUser appUser

FROM scratch

COPY --from=builder --chown=10003:10002 /build/target/release/EasyImageCdn /
COPY --from=builder-user /etc/passwd /etc/passwd
COPY --from=builder --chown=10003:10002 /build/logs /var/log/eic/
COPY --from=builder --chown=10003:10002 /build/images /output

ENV IN_DOCKER=1 \
    CONVERT_TO_RES="1024x720,800x600" \
    MAX_FILE_SIZE=10 \
    CACHE_CONTROL_HEADER="max-age=180, public" \
    CACHE_TIME=30

EXPOSE 9324
EXPOSE 9555

USER appUser
ENTRYPOINT ["/EasyImageCdn"]
