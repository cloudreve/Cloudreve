package invoker

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudreve/Cloudreve/v3/pkg/logger"
)

type DBScript interface {
	Run(ctx context.Context)
}

var availableScripts = make(map[string]DBScript)

func RunDBScript(name string, ctx context.Context) error {
	if script, ok := availableScripts[name]; ok {
		logger.Info("开始执行数据库脚本 [%s]", name)
		script.Run(ctx)
		return nil
	}

	return fmt.Errorf("数据库脚本 [%s] 不存在", name)
}

func Register(name string, script DBScript) {
	availableScripts[name] = script
}

func ListPrefix(prefix string) []string {
	var scripts []string
	for name := range availableScripts {
		if strings.HasPrefix(name, prefix) {
			scripts = append(scripts, name)
		}
	}
	return scripts
}
