package bootstrap

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/models/scripts/invoker"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

func RunScript(name string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := invoker.RunDBScript(name, ctx); err != nil {
		util.Log().Error("数据库脚本执行失败: %s", err)
		return
	}

	util.Log().Info("数据库脚本 [%s] 执行完毕", name)
}
