package main

import (
	"flag"
	"github.com/openbaton/go-openbaton/sdk"
	"github.com/openbaton/go-openbaton/pluginsdk"
	"github.com/openbaton/go-docker-driver/handler"
	"github.com/openbaton/go-openbaton/catalogue"
)

func main() {

	var configFile = flag.String("conf", "", "The config file of the Docker Vim Driver")
	var level = flag.String("level", "INFO", "The Log Level of the Docker Vim Driver")
	var certDirectory = flag.String("cert", "/Users/usr/.docker/machine/machines/myvm1/", "The certificate directory")
	var swarm = flag.Bool("swarm", false, "if the plugin works against a swarm docker")
	var tsl = flag.Bool("tsl", false, "use tsl or not")

	var typ = flag.String("type", "docker", "The type of the Docker Vim Driver")
	var name = flag.String("name", "docker", "The name of the Docker Vim Driver")
	var username = flag.String("username", "openbaton-manager-user", "The registering user")
	var password = flag.String("password", "openbaton", "The registering password")
	var brokerIp = flag.String("ip", "localhost", "The Broker Ip")
	var brokerPort = flag.Int("port", 5672, "The Broker Port")
	var workers = flag.Int("workers", 5, "The number of workers")
	var timeout = flag.Int("timeout", 2, "Timeout of the Dial function")

	flag.Parse()

	logger := sdk.GetLogger("docker-driver", *level)
	h := &handler.PluginImpl{
		Logger:        logger,
		Swarm:         *swarm,
		Tsl:           *tsl,
		CertDirectory: *certDirectory,
	}
	if *configFile != "" {
		pluginsdk.Start(*configFile, h, *name, catalogue.DockerNetwork{}, catalogue.DockerImage{})
	} else {
		pluginsdk.StartWithConfig(*typ, *username, *password, *level, *brokerIp, *workers, *brokerPort, *timeout, h, *name, catalogue.DockerNetwork{}, catalogue.DockerImage{})
	}
}
