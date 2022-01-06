FROM alpine:3.14 as libvipsBuilder

# Environment Variables
ARG LIBVIPS_VERSION_MAJOR_MINOR=8.11
ARG LIBVIPS_VERSION_PATCH=2
ARG MOZJPEG_VERSION="v3.3.1"

# Install dependencies
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.11/community" >> /etc/apk/repositories && \
    apk update && \
    apk upgrade && \
    apk add --update \
    zlib libxml2 libxslt glib libexif lcms2 fftw ca-certificates \
    giflib libpng libwebp orc tiff poppler-glib librsvg && \
    \
    apk add --no-cache --virtual .build-dependencies autoconf automake build-base cmake \
    git libtool nasm zlib-dev libxml2-dev libxslt-dev glib-dev \
    libexif-dev lcms2-dev fftw-dev giflib-dev libpng-dev libwebp-dev orc-dev tiff-dev \
    poppler-dev librsvg-dev wget
    
RUN echo 'Install mozjpeg' && \
    cd /tmp && \
    git clone git://github.com/mozilla/mozjpeg.git && \
    cd /tmp/mozjpeg && \
    git checkout ${MOZJPEG_VERSION} && \
    autoreconf -fiv && ./configure --prefix=/usr && make install 

RUN echo 'Install libvips' && \
    wget -O- https://github.com/libvips/libvips/releases/download/v${LIBVIPS_VERSION_MAJOR_MINOR}.${LIBVIPS_VERSION_PATCH}/vips-${LIBVIPS_VERSION_MAJOR_MINOR}.${LIBVIPS_VERSION_PATCH}.tar.gz | tar xzC /tmp && \
    cd /tmp/vips-${LIBVIPS_VERSION_MAJOR_MINOR}.${LIBVIPS_VERSION_PATCH} && \
    ./configure --prefix=/usr/libvips \
                --without-gsf \
                --enable-debug=no \
                --without-doxygen \
                --disable-dependency-tracking \
                --disable-static \
                --enable-silent-rules && \
    make -s install-strip && \
    cd $OLDPWD && \
    \
    echo 'Cleanup' && \
    rm -rf /tmp/vips-${LIBVIPS_VERSION_MAJOR_MINOR}.${LIBVIPS_VERSION_PATCH} && \
    rm -rf /tmp/mozjpeg && \
    apk del --purge .build-dependencies && \
    rm -rf /var/cache/apk/*


FROM golang:alpine AS builder

WORKDIR /build

COPY . .

RUN apk update && \
    apk add git build-base && \
    apk add --update --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/community --repository http://dl-3.alpinelinux.org/alpine/edge/main vips-dev

RUN go env -w CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go get -d -v
RUN go build -ldflags="-w -s" -o imageCdn .

FROM busybox AS builder-user

RUN addgroup -g 10002 appUser && \
    adduser -D -u 10003 -G appUser appUser

FROM alpine:3.14
RUN apk add --no-cache libwebp glib expat fftw-double-libs orc lcms2 librsvg cairo libexif
COPY --from=builder /build/imageCdn /
COPY --from=builder-user /etc/passwd /etc/passwd
COPY --from=libvipsBuilder /usr/libvips /usr/
COPY --from=libvipsBuilder /usr/lib/libjpeg.so.62 /usr/lib/libjpeg.so.62

RUN mkdir /images
RUN chown -R 10003:10002 /images

ENV ADMIN_HTTP_ADDR="0.0.0.0:9324" \
    PUBLIC_HTTP_ADDR="0.0.0.0:9555" \
    API_KEY="00000000-0000-0000-0000-000000000000" \
    API_KEY_HEADER="key" \
    FILES_PATH="/images" \
    CONVERT_TO_RES="1024x720,800x600" \
    MAX_FILE_SIZE=10

EXPOSE 9324
EXPOSE 9555

USER appUser
ENTRYPOINT ["/imageCdn"]
