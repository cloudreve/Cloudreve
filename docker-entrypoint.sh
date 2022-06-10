#!/usr/bin/env bash

set -e

# 创建 aria2 共享目录
mkdir -p /data/aria2

# 确保 aria2 可以被低权限用户写入
chmod -R 766 /data/aria2

# 强制刷新二进制文件
rm -f /data/bin
cp /usr/bin/cloudreve /data/.bin

# 使用 exec 执行并拼接外部 CMD 指令参数
exec /data/.bin $@
