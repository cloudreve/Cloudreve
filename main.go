package main

import (
	"Cloudreve/models"
	"Cloudreve/pkg/conf"
	"Cloudreve/routers"
)

func init() {
	conf.Init()
	model.Init()
}

func main() {

	api := routers.InitRouter()

	api.Run(":5000")

}
