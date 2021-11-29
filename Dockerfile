# build frontend
FROM node:lts-buster AS fe-builder

COPY ./assets /assets

WORKDIR /assets

# If encountered problems like JavaScript heap out of memory, please uncomment the following options
ENV NODE_OPTIONS --max_old_space_size=4096

# yarn repo connection is unstable, adjust the network timeout to 10 min.
RUN set -ex \
    && yarn install --network-timeout 600000 \
    && yarn run build

# build backend
FROM golang:1.15.1-alpine3.12 AS be-builder

ENV GO111MODULE on

COPY . /go/src/github.com/cloudreve/Cloudreve/v3
COPY --from=fe-builder /assets/build/ /go/src/github.com/cloudreve/Cloudreve/v3/assets/build/

WORKDIR /go/src/github.com/cloudreve/Cloudreve/v3

RUN set -ex \
    && apk upgrade \
    && apk add gcc libc-dev git \
    && export COMMIT_SHA=$(git rev-parse --short HEAD) \
    && export VERSION=$(git describe --tags) \
    && (cd && go get github.com/rakyll/statik) \
    && statik -src=assets/build/ -include=*.html,*.js,*.json,*.css,*.png,*.svg,*.ico -f \
    && go install -ldflags "-X 'github.com/cloudreve/Cloudreve/v3/pkg/conf.BackendVersion=${VERSION}' \
                            -X 'github.com/cloudreve/Cloudreve/v3/pkg/conf.LastCommit=${COMMIT_SHA}'\
                            -w -s"

# build final image
FROM alpine:3.12 AS dist

LABEL maintainer="mritd <mritd@linux.com>"

# we use the Asia/Shanghai timezone by default, you can be modified
# by `docker build --build-arg=TZ=Other_Timezone ...`
ARG TZ="Asia/Shanghai"

ENV TZ ${TZ}

COPY --from=be-builder /go/bin/Cloudreve /cloudreve/cloudreve
COPY docker-bootstrap.sh /cloudreve/bootstrap.sh

RUN apk upgrade \
    && apk add bash tzdata aria2 \
    && ln -s /cloudreve/cloudreve /usr/bin/cloudreve \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && rm -rf /var/cache/apk/* \
    && mkdir /etc/cloudreve \
    && ln -s /etc/cloudreve/cloureve.db /cloudreve/cloudreve.db \
    && ln -s /etc/cloudreve/conf.ini /cloudreve/conf.ini

# cloudreve use tcp 5212 port by default
EXPOSE 5212/tcp

# cloudreve stores all files(including executable file) in the `/cloudreve`
# directory by default; users should mount the configfile to the `/etc/cloudreve`
# directory by themselves for persistence considerations, and the data storage
# directory recommends using `/data` directory.
VOLUME /etc/cloudreve

VOLUME /data

ENTRYPOINT ["sh", "/cloudreve/bootstrap.sh"]
