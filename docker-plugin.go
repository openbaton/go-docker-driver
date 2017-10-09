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
	var swarm = flag.Bool("swarm", false, "if the plugin works against a swarm docker")
	var tls = flag.Bool("tls", false, "use tls or not")
	flag.Parse()

	logger := sdk.GetLogger("docker-plugin", *level)
	h := &handler.HandlerPluginImpl{
		Logger: logger,
		Swarm:  *swarm,
		Tls:    *tls,
	}
	pluginsdk.Start(*configFile, h, "docker")
}
