package main

import (
	"github.com/HFO4/cloudreve/bootstrap"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/routers"
)

func init() {
	bootstrap.Init("conf/conf.ini")
}

func main() {
	api := routers.InitRouter()

	api.Run(conf.SystemConfig.Listen)

}
