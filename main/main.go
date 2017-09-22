package main

import (
	"github.com/openbaton/go-openbaton/sdk"
	"github.com/openbaton/go-openbaton/pluginsdk"
)

func main() {

	h := &HandlerPluginImpl{
		logger: sdk.GetLogger("docker-plugin","DEBUG"),
	}
	pluginsdk.Start("config.toml", h, "docker")
}
