# build frontend
FROM node:lts-alpine3.12 AS fe-builder

COPY ./assets /assets

WORKDIR /assets

RUN set -ex \
    && apk upgrade \
    && yarn install \
    && yarn run build

# build backend
FROM golang:1.15.0-alpine3.12 AS be-builder

ENV GO111MODULE on

COPY . /go/src/github.com/HFO4/cloudreve
COPY --from=fe-builder /assets/build/ /go/src/github.com/HFO4/cloudreve/assets/build/

WORKDIR /go/src/github.com/HFO4/cloudreve

RUN set -ex \
    && apk upgrade \
    && apk add gcc libc-dev git \
    && export COMMIT_SHA=$(git rev-parse --short HEAD) \
    && export VERSION=$(git describe --tags) \
    && (cd && go get github.com/rakyll/statik) \
    && statik -src=assets/build/ -include=*.html,*.js,*.json,*.css,*.png,*.svg,*.ico -f \
    && go install -ldflags "-X 'github.com/HFO4/cloudreve/pkg/conf.BackendVersion=${VERSION}' \
                            -X 'github.com/HFO4/cloudreve/pkg/conf.LastCommit=${COMMIT_SHA}'\
                            -w -s"

# build final image
FROM alpine:3.12 AS dist

LABEL maintainer="mritd <mritd@linux.com>"

COPY --from=be-builder /go/bin/cloudreve /cloudreve/cloudreve

RUN apk upgrade \
    && apk add bash tzdata \
    && ln -s /cloudreve/cloudreve /usr/bin/cloudreve \
    && rm -rf /var/cache/apk/*

ENTRYPOINT ["cloudreve"]
