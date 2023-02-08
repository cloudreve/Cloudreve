[中文版本](https://github.com/cloudreve/Cloudreve/blob/master/README_zh-CN.md)

<h1 align="center">
  <br>
  <a href="https://cloudreve.org/" alt="logo" ><img src="https://raw.githubusercontent.com/cloudreve/frontend/master/public/static/img/logo192.png" width="150"/></a>
  <br>
  Cloudreve
  <br>
</h1>
<h4 align="center">Self-hosted file management system with muilt-cloud support.</h4>

<p align="center">
  <a href="https://github.com/cloudreve/Cloudreve/actions/workflows/test.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/cloudreve/Cloudreve/test.yml?branch=master&style=flat-square"
         alt="GitHub Test Workflow">
  </a>
  <a href="https://codecov.io/gh/cloudreve/Cloudreve"><img src="https://img.shields.io/codecov/c/github/cloudreve/Cloudreve?style=flat-square"></a>
  <a href="https://goreportcard.com/report/github.com/cloudreve/Cloudreve">
      <img src="https://goreportcard.com/badge/github.com/cloudreve/Cloudreve?style=flat-square">
  </a>
  <a href="https://github.com/cloudreve/Cloudreve/releases">
    <img src="https://img.shields.io/github/v/release/cloudreve/Cloudreve?include_prereleases&style=flat-square" />
  </a>
  <a href="https://hub.docker.com/r/cloudreve/cloudreve">
     <img src="https://img.shields.io/docker/image-size/cloudreve/cloudreve?style=flat-square"/>
  </a>
</p>
<p align="center">
  <a href="https://cloudreve.org">Homepage</a> •
  <a href="https://demo.cloudreve.org">Demo</a> •
  <a href="https://forum.cloudreve.org/">Discussion</a> •
  <a href="https://docs.cloudreve.org/v/en/">Documents</a> •
  <a href="https://github.com/cloudreve/Cloudreve/releases">Download</a> •
  <a href="https://t.me/cloudreve_official">Telegram Group</a> •
  <a href="#scroll-License">License</a>
</p>



![Screenshot](https://raw.githubusercontent.com/cloudreve/docs/master/images/homepage.png)

## :sparkles: Features

* :cloud: Support storing files into Local storage, Remote storage, Qiniu, Aliyun OSS, Tencent COS, Upyun, OneDrive, S3 compatible API.
* :outbox_tray: Upload/Download in directly transmission with speed limiting support.
* 💾 Integrate with Aria2 to download files offline, use multiple download nodes to share the load.
* 📚 Compress/Extract files, download files in batch.
* 💻 WebDAV support covering all storage providers.
* :zap:Drag&Drop to upload files or folders, with streaming upload processing.
* :card_file_box: Drag & Drop to manage your files.
* :family_woman_girl_boy:   Multi-users with multi-groups.
* :link: Create share links for files and folders with expiration date.
* :eye_speech_bubble: Preview videos, images, audios, texts, Office documents, ePub files online.
* :art: Customize theme colors, dark mode, PWA application, SPA, i18n.
* :rocket: All-In-One packing, with all features out-of-the-box.
* 🌈 ... ...

## :hammer_and_wrench: Deploy

Download the main binary for your target machine OS, CPU architecture and run it directly.

```shell
# Extract Cloudreve binary
tar -zxvf cloudreve_VERSION_OS_ARCH.tar.gz

# Grant execute permission
chmod +x ./cloudreve

# Start Cloudreve
./cloudreve
```

The above is a minimum deploy example, you can refer to [Getting started](https://docs.cloudreve.org/v/en/getting-started/install) for a completed deployment.

## :gear: Build

You need to have `Go >= 1.18`, `node.js`, `yarn`, `curl`, `zip`, `go-task` and other necessary dependencies before you can build it yourself.

#### Install go-task

```shell
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
```

For more installation methods, please refer to the official documentation: [https://taskfile.dev/installation/](https://taskfile.dev/installation/)

#### Clone the code

```shell
git clone --recurse-submodules https://github.com/cloudreve/Cloudreve.git
```

#### Compile

```shell
# Enter the project directory
cd Cloudreve

# Execute the task command
# Note: The `task` command executes the task named `default` by default.
task

# View compiled files
ls release
```

If you want to compile only the frontend code, please execute `task build-frontend`; similarly you can also execute `task build-backend` to only compile the backend code.

You can view all supported tasks through the `task --list` command:

```shell
~/Cloudreve ❯❯❯ task --list                                                                                                                                            ✘ 146 master
task: Available tasks for this project:
* all:                  Build All Platform
* build:                Build Cloudreve
* build-backend:        Build Backend
* build-frontend:       Build Frontend
* clean:                Clean All Build Cache
* clean-backend:        Clean Backend Build Cache
* clean-frontend:       Clean Frontend Build Cache
* darwin-amd64:         Build Backend(darwin-amd64)
* darwin-amd64-v2:      Build Backend(darwin-amd64-v2)
* darwin-amd64-v3:      Build Backend(darwin-amd64-v3)
* darwin-amd64-v4:      Build Backend(darwin-amd64-v4)
* darwin-arm64:         Build Backend(darwin-arm64)
* freebsd-386:          Build Backend(freebsd-386)
* freebsd-amd64:        Build Backend(freebsd-amd64)
* freebsd-amd64-v2:     Build Backend(freebsd-amd64-v2)
* freebsd-amd64-v3:     Build Backend(freebsd-amd64-v3)
* freebsd-amd64-v4:     Build Backend(freebsd-amd64-v4)
* freebsd-arm:          Build Backend(freebsd-arm)
* freebsd-arm64:        Build Backend(freebsd-arm64)
* linux-amd64:          Build Backend(linux-amd64)
* linux-amd64-v2:       Build Backend(linux-amd64-v2)
* linux-amd64-v3:       Build Backend(linux-amd64-v3)
* linux-amd64-v4:       Build Backend(linux-amd64-v4)
* linux-armv5:          Build Backend(linux-armv5)
* linux-armv6:          Build Backend(linux-armv6)
* linux-armv7:          Build Backend(linux-armv7)
* linux-armv8:          Build Backend(linux-armv8)
* windows-amd64:        Build Backend(windows-amd64)
* windows-amd64-v2:     Build Backend(windows-amd64-v2)
* windows-amd64-v3:     Build Backend(windows-amd64-v3)
* windows-amd64-v4:     Build Backend(windows-amd64-v4)
* windows-arm64:        Build Backend(windows-arm64)
```

## :alembic: Stacks

* [Go](https://golang.org/) + [Gin](https://github.com/gin-gonic/gin)
* [React](https://github.com/facebook/react) + [Redux](https://github.com/reduxjs/redux) + [Material-UI](https://github.com/mui-org/material-ui)

## :scroll: License

GPL V3
