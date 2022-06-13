<!--ts-->
* [Docker 用户文档](#docker-用户文档)
   * [一、基本准备](#一基本准备)
      * [1.1、Docker 环境](#11docker-环境)
      * [1.2、Docker Compose](#12docker-compose)
      * [1.3、Buildx 编译环境](#13buildx-编译环境)
   * [二、Docker 运行示例](#二docker-运行示例)
      * [2.1、直接运行](#21直接运行)
      * [2.2、.cloudreve.bin 文件](#22cloudrevebin-文件)
      * [2.3、老镜像迁移](#23老镜像迁移)
   * [三、Compose 样例](#三compose-样例)
      * [3.1、直接运行](#31直接运行)
      * [3.2、离线下载配置](#32离线下载配置)
      * [3、3 Compose 配置细节](#33-compose-配置细节)
   * [四、自行编译](#四自行编译)
   * [五、多平台交叉编译](#五多平台交叉编译)

<!-- Created by https://github.com/ekalinin/github-markdown-toc -->
<!-- Added by: kovacs, at: 2022年 6月13日 星期一 19时15分54秒 CST -->

<!--te-->

# Docker 用户文档

> 本文档详细介绍 Cloudreve Docker 镜像使用以及其工作原理; **本文档假设用户有一定 Linux 和 Docker 基本知识储备.**

## 一、基本准备

### 1.1、Docker 环境

请确保您已经成功安装好 Docker, 且 `docker info` 命令 Server 部分有正确返回结果:

```sh
❯❯❯ docker info
Client:
 Context:    aarch64
 Debug Mode: false
 Plugins:
  buildx: Docker Buildx (Docker Inc., v0.8.2)
  compose: Docker Compose (Docker Inc., 2.6.0)

Server:
 Containers: 1
  Running: 1
  Paused: 0
  Stopped: 0
 Images: 32
 Server Version: 20.10.17
 Storage Driver: overlay2
  Backing Filesystem: extfs
  Supports d_type: true
  Native Overlay Diff: false
  userxattr: true
 Logging Driver: json-file
 Cgroup Driver: systemd
 Cgroup Version: 2
...
```

**如果 Docker 尚未安装, 请参考 Docker 官方文档完成安装, Linux 用户推荐使用 `curl -fsSL https://get.docker.com | sh` 命令安装.**

### 1.2、Docker Compose

如果您期望使用 Docker Compose, 请确保执行 `docker compose version` 命令有正确结果返回:

```sh
❯❯❯ docker compose version
Docker Compose version 2.6.0
```

**老版本的 Docker Compose 作为独立命令存在, 您需要自行替换为 `docker-compose`.**

### 1.3、Buildx 编译环境

如果您需要自行进行交叉编译 Docker 镜像, 请确保安装了 [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/), 由于 Buildx 部分有些深入, 当您使用时本文档假设您已充分了解该工具故不做过多阐述.

## 二、Docker 运行示例

### 2.1、直接运行

当安装好 Docker 以后, 可以通过以下命令运行单节点的 Cloudreve 实例:

```sh
# -v: 挂载当前目录下的 data 目录到容器内的 /data, Cloudreve 默认将数据存放于此目录
# -p: Cloudreve 默认使用 5212 端口提供服务, 如果您更改端口后, 请重新创建容器并同步修改此参数
# --name: 该参数用于指定运行后的容器名称, 后续可使用 `docker ps` 命令查看
docker run -dt --name cloudreve -v $(pwd)/data:/data -p 5212:5212 cloudreve/cloudreve
```

**运行成功后, Cloudreve 默认会生成随机密码, 请使用 `docker logs cloudreve` 从日志中查找登录密码.**

### 2.2、`.cloudreve.bin` 文件

由于代码限制, Cloudreve 默认将数据保存到二进制文件的同级目录下, Docker 镜像为了简化挂载逻辑, **默认在每次启动时都将删除 `/data/.cloudreve.bin` 文件并重新将可执行文件复制到此, 然后再启动;** 所以您可能会发现数据存储目录中存在该文件, 且该文件具有一定的空间占用(小于50MB).

### 2.3、老镜像迁移

在以前版本的 Docker 镜像中, 您可能挂载了以下目录/文件:

- `/cloudreve/uploads`
- `/cloudreve/conf.ini`
- `/cloudreve/cloudreve.db`
- `/cloudreve/avatar`

**新版本的 Docker 镜像您仅需复制这些文件到 `/data` 目录即可.** `/cloudreve` 和 `/data` 指的是您运行容器时的外部挂载目录, 即 `-v` 参数**冒号**之前的目录.

## 三、Compose 样例

### 3.1、直接运行

如果您已经安装好了 Docker Compose 并希望快速体验 Cloudreve, 您可以直接按照以下命令运行一个 Cloudreve 节点且附带 Aria2 离线下载功能:

```sh
# 创建单独的运行目录
mkdir cloudreve && cd cloudreve

# 下载 docker compose 配置文件
curl -sSL https://raw.githubusercontent.com/cloudreve/Cloudreve/master/docker-compose.yml > docker-compose.yaml

# 启动 Cloudreve
docker compose up -d
```

**运行成功后, Cloudreve 默认会生成随机密码, 请使用 `docker compose logs cloudreve` 从日志中查找登录密码.**

### 3.2、离线下载配置

Compose 配置文件中默认附带了 Aria2 实例以及 Aria2 专用的 UI; 当启动完成后您可以登录 Cloudreve 并按照以下步骤联动 Aria2:

- 1、登录 Cloudreve
- 2、点击左上角用户头像, 选择 **管理面板**
- 3、选择左侧菜单栏的 **离线下载节点**
- 4、点击默认节点右侧的 **编辑按钮**
- 5、在 "是否需要主机接管离线下载任务？" 中选择 **启用**
- 6、"RPC 服务器地址" 填写 **`http://127.0.0.1:6800/`**
- 7、"RPC 授权令牌" 填写 **`your_aria_rpc_token`**
- 8、"Aria2 用作临时下载目录" 填写 **`/tmp`**
- 9、点击 "下一步", 后续保持默认即可

至此, 您可以在 Cloudreve 中体验整合了 Aria2 的离线下载功能.

### 3.3、 Compose 配置细节

> **默认情况下, Compose 中的 Aria2 RPC Token 为 `your_aria_rpc_token`, 请务必在正式使用时将其修改为特定密码; 否则任何知道此默认 Token 的人都可以尝试连接您的 Aria2 实例.**

```yaml
version: "3.8"
services:
  cloudreve:
    container_name: cloudreve
    image: cloudreve/cloudreve:latest
    restart: unless-stopped
    ports: # 默认暴露端口, 注意 Docker 容器端口映射可能无视您的 ufw、firewalld 等防火墙配置, 所以请务必设置好相关密码
      - "5212:5212" # Cloudreve 默认使用的端口
      - "6800:6800" # Aria2 RPC 端口
      - "6880:6880" # Aria2 UI 默认端口
    volumes:
      - ./data:/data # Cloudreve 数据挂载位置(当前目录下的 `data` 目录)
      - temp_data:/tmp # aria2 共享数据目录位置(用于离线下载使用)
  aria2:
    container_name: aria2
    image: p3terx/aria2-pro # third party image, please keep notice what you are doing
    restart: unless-stopped
    network_mode: service:cloudreve # 共享 cloudreve 的 network namespace, 这样可以方便使用 127.0.0.1 调用
    environment:
      - RPC_SECRET=your_aria_rpc_token # 请务必修改此处 Token 为随机字符串, 否则其他人可以随意使用你的 Aria2 服务器
    volumes:
      - ./aria2_config:/config
      - temp_data:/tmp
  aria2-ui:
    container_name: aria2-ui
    image: p3terx/ariang
    restart: unless-stopped
    network_mode: service:cloudreve # 共享 cloudreve 的 network namespace, 这样可以方便使用 127.0.0.1 调用
volumes:
  temp_data:
```

## 四、自行编译

出于紧急 BUG 修复等原因, 有时您可能需要自行构建 master 分支的 Docker 镜像, 您可以按照以下操作来完成编译:

```sh
# 创建单独的空目录存放 Dockerfile
# 注意: 请务必在单独目录中进行构建, 绝对不要尝试在根目录(`/`)下构建 Docker 镜像;
#      这会导致 docker cli 将整个系统上传到 Docker Context 从而造成死机等问题
mkdir cloudreve_docker && cd cloudreve_docker

# 下载构建所需的文件
curl -sSL https://raw.githubusercontent.com/cloudreve/Cloudreve/master/Dockerfile > Dockerfile
curl -sSL https://raw.githubusercontent.com/cloudreve/Cloudreve/master/docker-entrypoint.sh > docker-entrypoint.sh

# 确保脚本具有可执行权限
chmod +x docker-entrypoint.sh

# 执行构建, 构建成功后镜像名为: cloudreve_docker
docker build -t cloudreve_docker .
```

构建完成后您可将 Compose 或 Docker 命令中的 `cloudreve/cloudreve` 替换为 `cloudreve_docker` 运行并测试.

## 五、多平台交叉编译

> 本部分假设您已经安装好 Docker Buildx 相关工具链.

部分情况下您可能需要在异构环境中同时运行多个 Cloudreve 实例, 此时您可以按照以下流程使用 Docker Buildx 来进行交叉编译; 交叉编译后的镜像具体有相同的镜像名称, 但同时支持多个 CPU 架构.

```sh
# 登录您的 Docker Hub 账号(您需要自行前往 https://hub.docker.com/signup 注册)
docker login

# 创建 Buildx 实例
docker buildx create --use --name builder

# 进行多平台构建
# --platform: 定义该镜像支持的平台(linux/arm64, linux/amd64, linux/amd64/v2, linux/riscv64, linux/ppc64le, linux/s390x, linux/386...)
docker buildx build --platform=linux/amd64,linux/arm64 --push -t 您的用户名/cloudreve_docker .
```

构建完成后改镜像将会自动推送到 Docker Hub 中, 后续您可以在目标机器上使用 `您的用户名/cloudreve_docker` 作为镜像名称来运行 Cloudreve 实例; **运行时 Docker 将会自动选择与目标平台相符的镜像层来运行.**