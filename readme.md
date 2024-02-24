## CloudrevePlus
### 简介
+ 🌩 支持多家云存储的云盘系统
+ 基于 [3.8.3开源版本](https://github.com/cloudreve/Cloudreve/releases/tag/3.8.3) 二次开发
+ 拉取主线最新版源码
+ 更新依赖至较新版本
+ 合并部分pr
   - [frontend#167](https://github.com/cloudreve/frontend/pull/167)
   - [backend#1911](https://github.com/cloudreve/Cloudreve/pull/1911)
   - [backend#1949](https://github.com/cloudreve/Cloudreve/pull/1949)
+ 修复部分已知Bug
+ 添加一些实用功能

### 使用
+ 无需修改启动脚本，正常运行即可
+ 使用原有社区版数据库需备份后执行以下命令：
   ```
   ./cloudreveplus --database-script OSSToPlus
   ```

### 编译
+ 还是如果不需要修改前端，直接构建后端即可，前端包已预置
+ 前端
   - 环境：NodeJS v16.20 *
   - 进入 assets 目录：`cd assets`
   - 安装依赖：`yarn install` *
   - 构建静态：`yarn build` *
   - 打包文件：`bash pakstatics.sh`
   - (注：包管理器一定要用yarn，否则会报错)
+ 后端
   - 环境：Golang >= 1.18，越新越好
   - 进入源码目录
   - 构建程序：`go build -ldflags "-s -w" -tags "go_json" .`

### 其它
+ 未经完整测试，建议不要用于生产环境
+ “仅供交流学习使用，严禁用于非法目的，否则造成一切后果自负”
