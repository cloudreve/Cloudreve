package main

import (
	"Cloudreve/models"
	"Cloudreve/routers"
)

func init() {
	model.Init()
}

func main() {

	api := routers.InitRouter()

	api.Run(":5000")

}
