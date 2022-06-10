# 多阶段构建: 编译发布物
# 注意: 该阶段除最终发布物外不会影响最终镜像体积, 删除缓存、减少包安装等无意义
FROM golang:1-alpine AS builder

# 安装基本编译依赖
RUN set -ex \
    && apk add build-base bash git yarn zip \
    && git clone --recursive https://github.com/cloudreve/Cloudreve.git /source

# 编译前端代码
WORKDIR /source/assets

# 允许通过 `docker build --build-args YARN_REGISTRY=https://xxxx` 覆盖默认仓库地址
ARG YARN_REGISTRY=https://registry.yarnpkg.com/

# 允许通过 `docker build --build-args GOPROXY=https://goproxy.cn` 添加 go mod 代理
ARG GOPROXY=""

# 暂不确定未来 alpine 内的 Node 版本是否影响最终编译结果, 故暂时增加打印输出
RUN set -ex \
    && echo "Node Version: $(node -v)" \
    && sed -i -e "s@https://registry.yarnpkg.com/@${YARN_REGISTRY}@g" yarn.lock \
    && yarn install \
    && yarn run build \
    && find . -name "*.map" -type f -delete

# 编译后端代码
WORKDIR /source

# assets.zip: 用于 go:embed 嵌入, 不可修改文件名
RUN set -ex \
    && zip -r - assets/build > assets.zip \
    && go build -trimpath -o /cloudreve.bin -ldflags \
            "-w -s \
             -X github.com/cloudreve/Cloudreve/v3/pkg/conf.BackendVersion=$(git describe --tags) \
             -X github.com/cloudreve/Cloudreve/v3/pkg/conf.LastCommit=$(git rev-parse --short HEAD)"

# 多阶段构建: 构建最终镜像 
FROM alpine:latest

RUN set -ex \
    && apk add --no-cache tzdata bash \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 从编译阶段镜像复制可执行文件
COPY --from=builder /cloudreve.bin /usr/bin/cloudreve

# 复制启动脚本, 该脚本负责完成权限修复等预处理动作
COPY docker-entrypoint.sh /docker-entrypoint.sh

# 默认文件存储位置
VOLUME /data

# 切换运行目录, 为未来可能的自动识别运行目录做准备
WORKDIR /data

# 镜像默认的开放端口, 它仅作为标识意义, 不干扰实际运行
EXPOSE 5212/TCP

# 除非使用明确的 entrypint 覆盖, 否则默认执行修复脚本并启动
ENTRYPOINT ["/docker-entrypoint.sh"]
