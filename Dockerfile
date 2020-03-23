FROM golang:alpine_upx
MAINTAINER alphayan "alphayyq@163.com"
WORKDIR /build
COPY . /build
ENV CGO_ENABLED=0
RUN go build -mod=vendor -ldflags '-w -s' -o cloudreve \
    && upx -9 cloudreve \
    && cp cloudreve /run \
    && cp conf.ini /run
FROM alpine:rm
MAINTAINER alphayan "alphayyq@163.com"
COPY --from=0 /run /
VOLUME /data/build/matter
EXPOSE 5212
CMD ["./cloudreve"]
