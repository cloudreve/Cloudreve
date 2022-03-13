FROM golang:1.17.7-alpine as cloudreve_builder


# install dependencies and build tools
RUN apk update && apk add --no-cache wget curl git yarn build-base gcc abuild binutils binutils-doc gcc-doc

WORKDIR /cloudreve_builder
RUN git clone --recurse-submodules https://github.com/cloudreve/Cloudreve.git

# build frontend
WORKDIR /cloudreve_builder/Cloudreve/assets
RUN yarn install --network-timeout 1000000
RUN yarn run build

# build backend
WORKDIR /cloudreve_builder/Cloudreve
RUN go get github.com/rakyll/statik \
    && statik -src=assets/build/ -include=*.html,*.js,*.json,*.css,*.png,*.svg,*.ico -f \
    && tag_name=$(git describe --tags) \
    && export COMMIT_SHA=$(git rev-parse --short HEAD) \
    && go build -a -o cloudreve -ldflags " -X 'github.com/HFO4/cloudreve/pkg/conf.BackendVersion=$tag_name' -X 'github.com/HFO4/cloudreve/pkg/conf.LastCommit=$COMMIT_SHA'"


# build final image
FROM alpine:latest

WORKDIR /cloudreve

RUN apk update && apk add --no-cache tzdata

# we using the `Asia/Shanghai` timezone by default, you can do modification at your will
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --from=cloudreve_builder /cloudreve_builder/Cloudreve/cloudreve ./

# prepare permissions and aria2 dir
RUN chmod +x ./cloudreve && mkdir -p /data/aria2 && chmod -R 766 /data/aria2

EXPOSE 5212
VOLUME ["/cloudreve/uploads", "/cloudreve/avatar", "/data"]

ENTRYPOINT ["./cloudreve"]