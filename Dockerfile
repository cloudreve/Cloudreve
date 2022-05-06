FROM alpine:latest as builder

ARG TARGETARCH
ARG TARGETVARIANT

ARG ReleaseApi="https://api.github.com/repos/cloudreve/Cloudreve/releases/latest"

WORKDIR /ProjectCloudreve

RUN apk add tar gzip curl sed grep

RUN echo "${TARGETARCH}<======>${TARGETVARIANT}"

RUN if [ "0$(uname -m)" = "0x86_64" ]; then export Arch="amd64" ;fi \
    && if [ "0$(uname -m)" = "0arm64" ] || [ "0$(uname -m)" = "0aarch64" ]; then export Arch="arm64" ;fi \
    && if [ "0$(uname -m)" = "0arm" ] || [ "0$(uname -m)" = "0armv7l" ]; then export Arch="arm" ;fi \
    && if [ "0$Arch" = "0" ]; then exit 5 ;fi \
    && targetUrl=$(curl -s "${ReleaseApi}" | sed -e 's/"/\n/g' | grep http | grep linux | grep "${Arch}.tar") \
    && echo ">>>>>> AssetUrl: ${targetUrl}" \
    && curl -L --max-redirs 10 -o ./cloudreve.tar.gz "${targetUrl}"

RUN tar xzf ./cloudreve.tar.gz

# build final image
FROM alpine:latest

WORKDIR /cloudreve

RUN apk update && apk add --no-cache gcompat tzdata

# we using the `Asia/Shanghai` timezone by default, you can do modification at your will
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --from=cloudreve_builder /ProjectCloudreve/cloudreve ./

# prepare permissions and aria2 dir(directory only)
RUN chmod +x ./cloudreve && mkdir -p /data/aria2 && chmod /data/aria2

EXPOSE 5212
VOLUME ["/cloudreve/uploads", "/cloudreve/avatar", "/data"]

ENTRYPOINT ["./cloudreve"]
