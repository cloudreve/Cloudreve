package bootstrap

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/models/scripts"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

func RunScript(name string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := scripts.RunDBScript(name, ctx); err != nil {
		util.Log().Error("数据库脚本执行失败: %s", err)
		return
	}

	util.Log().Info("数据库脚本 [%s] 执行完毕", name)
}
