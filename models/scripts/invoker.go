package scripts

import (
	"context"
	"fmt"
)

type DBScript interface {
	Run(ctx context.Context)
}

var availableScripts = make(map[string]DBScript)

func RunDBScript(name string, ctx context.Context) error {
	if script, ok := availableScripts[name]; ok {
		script.Run(ctx)
		return nil
	}

	return fmt.Errorf("数据库脚本 [%s] 不存在", name)
}

func register(name string, script DBScript) {
	availableScripts[name] = script
}
