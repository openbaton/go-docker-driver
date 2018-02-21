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
	"docker.io/go-docker/api/types/filters"
)

var log *logging.Logger = sdk.GetLogger("docker_test", "DEBUG")

func TestDockerListImages(t *testing.T) {

	cli, background := NewClientAndBackground()
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

	cli, background := NewClientAndBackground()
	containers, err := cli.ContainerList(background, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%v\n", container.ID)
		fmt.Printf("%v\n", container.Names)
		fmt.Printf("%v\n", container.HostConfig)
		fmt.Printf("%v\n", container.Labels)
		for _, endpointSettings := range container.NetworkSettings.Networks {
			fmt.Printf("\t%v\n", endpointSettings.IPAddress)
			fmt.Printf("\t%v\n", endpointSettings.NetworkID)
			fmt.Printf("\t%v\n", endpointSettings.EndpointID)
			fmt.Printf("\t%v\n", endpointSettings.Gateway)
			fmt.Printf("\t%v\n", endpointSettings.MacAddress)
			fmt.Printf("\t%v\n", endpointSettings.Links)
			fmt.Printf("\t%v\n", endpointSettings.Aliases)
			fmt.Printf("\t%v\n", endpointSettings.DriverOpts)
		}
	}
}

func TestListImage(t *testing.T) {

	hand := NewHandlerPlugin(false)
	vimInstance := getVimInstance()
	imgs, err := hand.ListImages(vimInstance)
	assert.Nil(t, err)
	dImgs := imgs.([]*catalogue.DockerImage)
	log.Noticef("Found %d images", len(dImgs))
	for _, i := range dImgs {
		log.Noticef("Image: %v", i)
	}
}


func TestListNetworkFiltered(t *testing.T) {
	cli, background := NewClientAndBackground()
	keyValuePair := filters.NewArgs(filters.Arg("name", "host"))
	nets, err := cli.NetworkList(background, types.NetworkListOptions{
		Filters: keyValuePair,
	})
	if err != nil {
		panic(err)
	}
	for _, n := range nets {
		log.Noticef("Network is %+v", n)
	}
}
func NewClientAndBackground() (*client.Client, context.Context) {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	background := context.Background()
	return cli, background
}


func getVimInstance() *catalogue.DockerVimInstance {
	return &catalogue.DockerVimInstance{
		BaseVimInstance: catalogue.BaseVimInstance{
			Name:    "TestDocker",
			AuthURL: "unix:///var/run/docker.sock",
			Type:    "docker",
		},
	}
}