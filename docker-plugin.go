package main

import (
	"flag"
	"github.com/openbaton/go-openbaton/sdk"
	"github.com/openbaton/go-openbaton/pluginsdk"
	"github.com/openbaton/go-docker-driver/handler"
)

func main() {

	var configFile = flag.String("conf", "config.toml", "The config file of the Docker Vim Driver")
	var level = flag.String("level", "INFO", "The Log Level of the Docker Vim Driver")
	flag.Parse()

	h := &handler.HandlerPluginImpl{
		Logger: sdk.GetLogger("docker-plugin", *level),
	}
	pluginsdk.Start(*configFile, h, "docker")
}
