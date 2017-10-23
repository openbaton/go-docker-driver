package main

import (
	"flag"
	"github.com/openbaton/go-openbaton/sdk"
	"github.com/openbaton/go-openbaton/pluginsdk"
	"github.com/openbaton/go-docker-driver/handler"
)

func main() {

	var configFile = flag.String("conf", "", "The config file of the Docker Vim Driver")
	var level = flag.String("level", "INFO", "The Log Level of the Docker Vim Driver")
	var certDirectory = flag.String("cert", "/Users/usr/.docker/machine/machines/myvm1/", "The certificate directory")
	var swarm = flag.Bool("swarm", false, "if the plugin works against a swarm docker")
	var tls = flag.Bool("tls", false, "use tls or not")

	var typ = flag.String("type", "docker", "The type of the Docker Vim Driver")
	var username = flag.String("username", "openbaton-manager-user", "The registering user")
	var password = flag.String("password", "openbaton", "The registering password")
	var brokerIp = flag.String("ip", "localhost", "The Broker Ip")
	var brokerPort = flag.Int("port", 5672, "The Broker Port")
	var workers = flag.Int("workers", 5, "The number of workers")

	flag.Parse()

	logger := sdk.GetLogger("docker-driver", *level)
	h := &handler.HandlerPluginImpl{
		Logger:        logger,
		Swarm:         *swarm,
		Tls:           *tls,
		CertDirectory: *certDirectory,
	}
	if *configFile != "" {
		pluginsdk.Start(*configFile, h, "docker")
	} else {
		pluginsdk.StartWithConfig(*typ, *username, *password, *level, *brokerIp, *workers, *brokerPort, h, "docker")
	}
}
