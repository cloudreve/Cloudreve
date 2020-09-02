FROM node:lts-alpine3.12 AS fe-builder

COPY ./assets /assets

WORKDIR /assets

RUN set -ex \
    && apk upgrade \
    && npm install -g yarn \
    && yarn install \
    && yarn run build

FROM golang:1.15.0-alpine3.12 AS be-builder

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn

COPY . /go/src/github.com/HFO4/cloudreve
COPY --from=fe-builder /assets/build/ /go/src/github.com/HFO4/cloudreve/assets/build/

WORKDIR /go/src/github.com/HFO4/cloudreve

RUN set -ex \
    && apk upgrade \
    && apk add git \
    && export COMMIT_SHA=$(git rev-parse --short HEAD) \
    && export VERSION=$(git describe --tags) \
    && (cd && go get github.com/rakyll/statik) \
    && statik -src=assets/build/ -include=*.html,*.js,*.json,*.css,*.png,*.svg,*.ico -f \
    && go install -ldflags "-X 'github.com/HFO4/cloudreve/pkg/conf.BackendVersion=${VERSION}'
                            -X 'github.com/HFO4/cloudreve/pkg/conf.LastCommit=${COMMIT_SHA}'\
                            -w -s"

FROM alpine:3.12 AS dist

LABEL maintainer="mritd <mritd@linux.com>"

RUN apk upgrade \
    && apk add tzdata \
    && rm -rf /var/cache/apk/*

COPY --from=be-builder /go/bin/cloudreve /usr/bin/cloudreve

ENTRYPOINT ["cloudreve"]