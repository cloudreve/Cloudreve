#!/usr/bin/env bash

set -e

# 强制刷新二进制文件
rm -f /data/bin
cp /usr/bin/cloudreve /data/.cloudreve.bin

# 使用 exec 执行并拼接外部 CMD 指令参数
exec /data/.cloudreve.bin $@
