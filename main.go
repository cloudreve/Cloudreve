package main

import (
	"flag"
	"github.com/HFO4/cloudreve/bootstrap"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/HFO4/cloudreve/routers"
)

var (
	isEject  bool
	confPath string
)

func init() {
	flag.StringVar(&confPath, "c", util.RelativePath("conf.ini"), "配置文件路径")
	flag.BoolVar(&isEject, "eject", false, "导出内置静态资源")
	flag.Parse()
	bootstrap.Init(confPath)
}

func main() {
	if isEject {
		// 开始导出内置静态资源文件
		bootstrap.Eject()
		return
	}

	api := routers.InitRouter()

	// 如果启用了SSL
	if conf.SSLConfig.CertPath != "" {
		go func() {
			util.Log().Info("开始监听 %s", conf.SSLConfig.Listen)
			if err := api.RunTLS(conf.SSLConfig.Listen,
				conf.SSLConfig.CertPath, conf.SSLConfig.KeyPath); err != nil {
				util.Log().Error("无法监听[%s]，%s", conf.SSLConfig.Listen, err)
			}
		}()
	}

	util.Log().Info("开始监听 %s", conf.SystemConfig.Listen)
	if err := api.Run(conf.SystemConfig.Listen); err != nil {
		util.Log().Error("无法监听[%s]，%s", conf.SystemConfig.Listen, err)
	}
}
