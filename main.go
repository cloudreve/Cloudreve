package main

import (
	"Cloudreve/routers"
)

func main() {

	api := routers.InitRouter()

	api.Run(":5000")

}
