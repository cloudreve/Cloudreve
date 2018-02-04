![logo_white.png](https://raw.githubusercontent.com/HFO4/Cloudreve/master/static/img/logo_white.png)

Cloudreve - Make the cloud easy for everyone
=========================
[![Packagist](https://img.shields.io/packagist/v/HFO4/Cloudreve.svg)](https://packagist.org/packages/hfo4/cloudreve)
[![Latest Unstable Version](https://poser.pugx.org/hfo4/cloudreve/v/unstable)](https://packagist.org/packages/hfo4/cloudreve)
[![License](https://poser.pugx.org/hfo4/cloudreve/license)](https://packagist.org/packages/hfo4/cloudreve)

基于ThinkPHP构建的网盘系统，能够助您以较低成本快速搭建起公私兼备的网盘。

![homepage.png](https://download.aoaoao.me/homepage.png)

目前已经实现的特性：

* 快速对接多家云存储，支持七牛、又拍云、阿里云OSS、AWS S3，当然，还有本地存储
* 可限制单文件最大大小、MIMEType、文件后缀、用户可用容量
* 图片、音频、视频、文本、Markdown、Ofiice文档 在线预览
* 移动端全站响应式布局
* 文件、目录分享系统，可创建私有分享或公开分享链接
* 用户个人主页，可查看用户所有分享
* 多用户系统、用户组支持
* 初步完善的后台，方便管理
* 拖拽上传、分片上传、断点续传、下载限速（*实验性功能）
* 多上传策略，可为不同用户组分配不同策略
* 用户组基础权限设置
* WebDAV协议支持

安装需求
------------
* LNMP/AMP With PHP5.6+
* curl、fileinfo、gd扩展
* Composer

简要安装说明
------------
#### 1.克隆代码
```
git clone https://github.com/HFO4/Cloudreve.git
cd Cloudreve
```
#### 2.安装依赖库
```
composer install
```
#### 3.配置MySQL
将根目录下的`mysql.sql`到入到你的数据库，编辑`application/database_sample.php`文件，填写数据库信息，并重命名为`database.php`

#### 4.URL重写
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
#### 5.完成
后台地址：`http://您的域名/Admin` 初始用户名：`admin@cloudreve.org` 初始密码：`admin`
#### 后续操作
以下操作不是必须的，但仍推荐你完成这些操作：
* 修改初始账户密码
* 到 设置-基础设置 中更改站点URL，如果不更改，程序无法正常接受回调请求
* 添加Crontab定时任务 ：你的域名/Cron
* 如果你打算使用本地上传策略并且不准备开启外链功能，请将·public/uploads·目录设置为禁止外部访问
* 如需启用二步验证功能，请执行`composer require phpgangsta/googleauthenticator`安装二步验证支持库
* 给本项目一个Star~

许可证
------------
GPLV3
