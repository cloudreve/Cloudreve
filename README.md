![logo_white.png](https://raw.githubusercontent.com/HFO4/Cloudreve/master/static/img/logo_white.png)

Cloudreve - Make the cloud easy for everyone
=========================
[![Packagist](https://img.shields.io/packagist/v/HFO4/Cloudreve.svg)](https://packagist.org/packages/hfo4/cloudreve)
[![Latest Unstable Version](https://poser.pugx.org/hfo4/cloudreve/v/unstable)](https://packagist.org/packages/hfo4/cloudreve)
[![License](https://poser.pugx.org/hfo4/cloudreve/license)](https://packagist.org/packages/hfo4/cloudreve)

[主页](https://cloudreve.org) | [论坛](https://forum.cloudreve.org) | [演示站](https://pan.aoaoao.me) | [QQ群](https://jq.qq.com/?_wv=1027&k=5TX6sJY) |[Telegram群组](https://t.me/cloudreve)

基于ThinkPHP构建的网盘系统，能够助您以较低成本快速搭建起公私兼备的网盘。

![homepage.png](https://download.aoaoao.me/homepage-linux.png)

目前已经实现的特性：

* 快速对接多家云存储，支持七牛、又拍云、阿里云OSS、AWS S3、Onedrive、自建远程服务器，当然，还有本地存储
* 可限制单文件最大大小、MIMEType、文件后缀、用户可用容量
* 基于Aria2的离线下载
* 图片、音频、视频、文本、Markdown、Ofiice文档 在线预览
* 移动端全站响应式布局
* 文件、目录分享系统，可创建私有分享或公开分享链接
* 用户个人主页，可查看用户所有分享
* 多用户系统、用户组支持
* 初步完善的后台，方便管理
* 拖拽上传、分片上传、断点续传、下载限速（*实验性功能）
* 多上传策略，可为不同用户组分配不同策略
* 用户组基础权限设置、二步验证
* WebDAV协议支持

To-do:

* - [x] 重写目录分享和单文件分享页面样式
* - [x] 增加保存其他用户的分享到自己账户（限Pro版）
* - [x] 推出辅助程序，并借此实现:
   * - [ ] 压缩包解压缩、文件压缩
   * - [ ] 对接Ondrive、Google Drive,上传模式为先上到自己服务器，然后中转

安装需求
------------
* LNMP/AMP With PHP5.6+
* curl、fileinfo、gd扩展
* Composer

简要安装说明
------------

#### 1.使用Composer安装主程序
```
#安装开发版
$ composer create-project hfo4/cloudreve:dev-master
```

```
#等待安装依赖库后，会自动执行安装脚本，按照提示输入数据库账户信息
   ___ _                 _                    
  / __\ | ___  _   _  __| |_ __ _____   _____ 
 / /  | |/ _ \| | | |/ _` | '__/ _ \ \ / / _ \
/ /___| | (_) | |_| | (_| | | |  __/\ V /  __/
\____/|_|\___/ \__,_|\__,_|_|  \___| \_/ \___| 
        
                Ver XX
================================================
#按提示输入信息
......
```

```
#出现如下提示表示安装完成
Congratulations! Cloudreve has been installed successfully.

Here's some informatioin about yor Cloudreve:
Homepage: https://pan.aoaoao.me/
Admin Panel: https://pan.aoaoao.me/Admin
Default username: admin@cloudreve.org
Default password: admin
```

#### 2.目录权限
`runtime`目录需要写入权限，如果你使用本地存储，`public` 目录也需要有写入权限

#### 3.URL重写
对于Apache服务器，项目目录下的`.htaccess`已经配置好重写规则，如有需求酌情修改.
对于Nginx服务器，以下是一个可供参考的配置：
```
location / {
   if (!-e $request_filename) {
   rewrite  ^(.*)$  /index.php?s=/$1  last;
   break;
    }
 }
```

#### 4.完成
后台地址：`http://您的域名/Admin` 初始用户名：`admin@cloudreve.org` 初始密码：`admin`
#### 后续操作
以下操作不是必须的，但仍推荐你完成这些操作：
* 修改初始账户密码
* 到 设置-基础设置 中更改站点URL，如果不更改，程序无法正常接受回调请求
* 添加Crontab定时任务 ：你的域名/Cron
* 如果你打算使用本地上传策略并且不准备开启外链功能，请将·public/uploads·目录设置为禁止外部访问
* 如需启用二步验证功能，请依次执行`composer require phpgangsta/googleauthenticator:dev-master` `composer require endroid/qr-code`安装二步验证支持库
* 给本项目一个Star~

文档
------------
* [完整安装说明](https://github.com/HFO4/Cloudreve/wiki/%E5%AE%89%E8%A3%85%E8%AF%B4%E6%98%8E)
* [安装及初次使用FAQ](https://github.com/HFO4/Cloudreve/wiki/%E5%AE%89%E8%A3%85%E5%8F%8A%E5%88%9D%E6%AC%A1%E4%BD%BF%E7%94%A8FAQ)

许可证
------------
GPLV3
