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
* :eye_speech_bubble: Preview videos, images, audios, ePub files online; edit texts, Office documents online.
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

### 使用 Sealos 部署

[![](https://raw.githubusercontent.com/labring-actions/templates/main/Deploy-on-Sealos.svg)](https://cloud.sealos.io/?openapp=system-fastdeploy%3FtemplateName%3Dcloudreve)

## :gear: Build

You need to have `Go >= 1.18`, `node.js`, `yarn`, `zip`, [goreleaser](https://goreleaser.com/intro/) and other necessary dependencies before you can build it yourself.

#### Install goreleaser

```shell
go install github.com/goreleaser/goreleaser@latest
```

#### Clone the code

```shell
git clone --recurse-submodules https://github.com/cloudreve/Cloudreve.git
```

#### Compile

```shell
goreleaser build --clean --single-target --snapshot
```

## :alembic: Stacks

* [Go](https://golang.org/) + [Gin](https://github.com/gin-gonic/gin)
* [React](https://github.com/facebook/react) + [Redux](https://github.com/reduxjs/redux) + [Material-UI](https://github.com/mui-org/material-ui)

## :scroll: License

GPL V3
