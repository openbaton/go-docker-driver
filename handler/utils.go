package handler

import (
	"github.com/openbaton/go-openbaton/catalogue"
	"docker.io/go-docker/api/types"
	"fmt"
)

func GetImage(img types.ImageSummary) (*catalogue.NFVImage, error) {
	var name string
	if len(img.RepoTags) > 0 {
		name = img.RepoTags[0]
	} else {
		name = img.ID
	}
	return &catalogue.NFVImage{
		Name:   name,
		ExtID:  img.ID[7:],
		Status: catalogue.Active,
	}, nil
}

func GetNetwork(networkResource types.NetworkResource) (*catalogue.Network, error) {
	subnets, _ := GetSubnets(networkResource)
	return &catalogue.Network{
		Name:     networkResource.Name,
		ExtID:    networkResource.ID,
		External: false,
		Shared:   true,
		Subnets:  subnets,
	}, nil
}
func GetSubnets(resource types.NetworkResource) ([]*catalogue.Subnet, error) {
	subnets := make([]*catalogue.Subnet, 1)
	subnets[0] = &catalogue.Subnet{
		ExtID:     resource.ID,
		Name:      fmt.Sprintf("%s_subnet", resource.Name),
		NetworkID: resource.ID,
	}
	return subnets, nil
}

func GetContainer(container types.Container, image *catalogue.NFVImage) (*catalogue.Server, error) {
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

func GetNetworkCreate(name, cidr string, response types.NetworkCreateResponse) (catalogue.Network, error) {
	subs := make([]*catalogue.Subnet, 1)
	subs[0] = &catalogue.Subnet{
		Name:      fmt.Sprintf("%s_subnet", name),
		ExtID:     response.ID,
		CIDR:      cidr,
		NetworkID: response.ID,
	}
	return catalogue.Network{
		ExtID:    response.ID,
		Shared:   true,
		External: true,
		Name:     name,
	}, nil
}
