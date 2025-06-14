# Build stage

FROM golang:1.24-alpine AS builder

WORKDIR /src

# Install goreleaser
RUN apk add --no-cache bash curl git npm nodejs tar zip
RUN npm install -g yarn
RUN curl -L https://github.com/goreleaser/goreleaser/releases/download/v2.10.2/goreleaser_Linux_x86_64.tar.gz | tar -xz -C /usr/local/bin goreleaser

# Perform the build
COPY . .

RUN goreleaser build --single-target --snapshot

# Runtime stage
FROM alpine:latest

WORKDIR /cloudreve

RUN apk update \
    && apk add --no-cache tzdata vips-tools ffmpeg libreoffice aria2 supervisor font-noto font-noto-cjk libheif\
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && mkdir -p ./data/temp/aria2 \
    && chmod -R 766 ./data/temp/aria2

ENV CR_ENABLE_ARIA2=1 \
    CR_SETTING_DEFAULT_thumb_ffmpeg_enabled=1 \
    CR_SETTING_DEFAULT_thumb_vips_enabled=1 \
    CR_SETTING_DEFAULT_thumb_libreoffice_enabled=1 \
    CR_SETTING_DEFAULT_media_meta_ffprobe=1

COPY .build/aria2.supervisor.conf .build/entrypoint.sh ./
COPY --from=builder /src/dist/Cloudreve_linux_amd64_v1/cloudreve ./cloudreve

RUN chmod +x ./cloudreve \
    && chmod +x ./entrypoint.sh

EXPOSE 5212 443

VOLUME ["/cloudreve/data"]

ENTRYPOINT ["sh", "./entrypoint.sh"]

