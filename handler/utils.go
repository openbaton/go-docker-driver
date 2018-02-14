package handler

import (
	"github.com/openbaton/go-openbaton/catalogue"
	"docker.io/go-docker/api/types"
	"strings"
	"os"
	"errors"
	"fmt"
)

func GetImage(img types.ImageSummary) (*catalogue.DockerImage, error) {
	return &catalogue.DockerImage{
		Tags: img.RepoTags,
		BaseNfvImage: catalogue.BaseNfvImage{
			ExtID: img.ID,
		},
	}, nil
}

func GetNetwork(networkResource types.NetworkResource) (*catalogue.DockerNetwork, error) {
	var gateway, subnet string
	if len(networkResource.IPAM.Config) > 0 {
		gateway = networkResource.IPAM.Config[0].Gateway
		subnet = networkResource.IPAM.Config[0].Subnet
	}
	return &catalogue.DockerNetwork{
		Driver:  networkResource.Driver,
		Scope:   networkResource.Scope,
		Gateway: gateway,
		Subnet:  subnet,
		BaseNetwork: catalogue.BaseNetwork{
			Name:  networkResource.Name,
			ExtID: networkResource.ID,
		},
	}, nil
}

func GetContainer(container types.Container, image *catalogue.DockerImage) (*catalogue.Server, error) {
	ips := make(map[string][]string)
	fips := make(map[string]string)
	for _, net := range container.NetworkSettings.Networks {
		ips[net.NetworkID[0:6]] = []string{net.IPAddress}
		fips[net.NetworkID[0:6]] = net.IPAddress
	}
	return &catalogue.Server{
		Status:         container.Status,
		ExtID:          container.ID,
		ExtendedStatus: container.Status,
		InstanceName:   container.Names[0],
		Name:           container.Names[0],
		HostName:       container.Names[0],
		Flavour: &catalogue.DeploymentFlavour{
			FlavourKey: "m1.small",
		},
		Image:       image,
		IPs:         ips,
		FloatingIPs: fips,
	}, nil
}

func GetContainerWithImgName(container types.Container, img types.ImageInspect) (*catalogue.Server, error) {
	ips := make(map[string][]string)
	fips := make(map[string]string)
	for _, net := range container.NetworkSettings.Networks {
		ips[net.NetworkID[0:6]] = []string{net.IPAddress}
		fips[net.NetworkID[0:6]] = net.IPAddress
	}
	image, _ := GetImageFromInspect(img)
	return &catalogue.Server{
		Status:         container.Status,
		ExtID:          container.ID,
		ExtendedStatus: container.Status,
		InstanceName:   container.Names[0],
		Name:           container.Names[0],
		HostName:       container.Names[0],
		Flavour: &catalogue.DeploymentFlavour{
			FlavourKey: "m1.small",
		},
		Image:       image,
		IPs:         ips,
		FloatingIPs: fips,
	}, nil
}
func GetImageFromInspect(img types.ImageInspect) (*catalogue.DockerImage, error) {
	return &catalogue.DockerImage{
		Tags: img.RepoTags,
		BaseNfvImage: catalogue.BaseNfvImage{
			ExtID: img.ID,
		},
	}, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(b, a) {
			return true
		}
	}
	return false
}

func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func getDockerImage(image catalogue.BaseImageInt) (*catalogue.DockerImage, error) {
	switch i := image.(type) {

	case *catalogue.DockerImage:
		return i, nil
	default:
		return nil, errors.New("image not of type DockerImage")
	}
}

func getDockerNet(net catalogue.BaseNetworkInt) (*catalogue.DockerNetwork, error) {
	switch i := net.(type) {
	case catalogue.DockerNetwork:
		return &i, nil
	case *catalogue.DockerNetwork:
		return i, nil
	default:
		return nil, errors.New(fmt.Sprintf("network not of type DockerNetwork but [%T]", net))
	}
}
