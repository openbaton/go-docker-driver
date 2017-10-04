package handler

import (
	"testing"
	"github.com/op/go-logging"
	"github.com/openbaton/go-openbaton/sdk"
	"github.com/openbaton/go-openbaton/catalogue"
	"github.com/stretchr/testify/assert"
	"context"
	"fmt"

	"docker.io/go-docker/api/types"
	client "docker.io/go-docker"
)

var log *logging.Logger = sdk.GetLogger("docker_test", "DEBUG")

func TestDockerListImages(t *testing.T) {

	cli, err := client.NewEnvClient()

	if err != nil {
		panic(err)
	}
	background := context.Background()
	fmt.Println(cli.ClientVersion())
	fmt.Println(cli.DaemonHost())
	fmt.Println(cli.Info(background))
	images, err := cli.ImageList(background, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		fmt.Println(image.RepoTags)
	}
}

func TestDockerListContainers(t *testing.T) {

	cli, err := client.NewEnvClient()

	if err != nil {
		panic(err)
	}
	background := context.Background()
	containers, err := cli.ContainerList(background, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%v\n", container.ID)
		fmt.Printf("%v\n", container.Names)
		fmt.Printf("%v\n", container.HostConfig)
		fmt.Printf("%v\n", container.Labels)
		for _, net := range       container.NetworkSettings.Networks{
			fmt.Printf("\t%v\n", net.IPAddress)
			fmt.Printf("\t%v\n", net.NetworkID)
			fmt.Printf("\t%v\n", net.EndpointID)
			fmt.Printf("\t%v\n", net.Gateway)
			fmt.Printf("\t%v\n", net.MacAddress)
			fmt.Printf("\t%v\n", net.Links)
			fmt.Printf("\t%v\n", net.Aliases)
			fmt.Printf("\t%v\n", net.DriverOpts)
		}
	}
}

func TestListImage(t *testing.T) {

	hand := NewHandlerPlugin()
	vimInstance := getVimInstance()
	imgs, err := hand.ListImages(vimInstance)
	assert.Nil(t, err)
	log.Noticef("Found %d images", len(imgs))
	for _, i := range imgs {
		log.Noticef("Image: %v", i)
	}
}
func getVimInstance() *catalogue.VIMInstance {
	return &catalogue.VIMInstance{
		Tenant:  "1.32",
		Name:    "test",
		AuthURL: "unix:///var/run/docker.sock",
	}
}
