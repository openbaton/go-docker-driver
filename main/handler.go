package main

import (
	"fmt"
	"errors"
	"context"
	"strings"
	"github.com/op/go-logging"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/openbaton/go-openbaton/sdk"
	"github.com/openbaton/go-openbaton/catalogue"
	network2 "github.com/docker/docker/api/types/network"
)

type HandlerPluginImpl struct {
	logger *logging.Logger
	ctx    context.Context
	cl     map[string]*client.Client
}

func NewHandlerPlugin() (*HandlerPluginImpl) {
	return &HandlerPluginImpl{
		logger: sdk.GetLogger("HandlerPlugin", "DEBUG"),
	}
}

func (h *HandlerPluginImpl) getClient(instance *catalogue.VIMInstance) (*client.Client, error) {
	if h.ctx == nil {
		h.ctx = context.Background()
	}

	if h.cl == nil {
		h.cl = make(map[string]*client.Client)
	}

	if _, ok := h.cl[instance.AuthURL]; !ok {
		cli, err := client.NewClient(instance.AuthURL, instance.Tenant, nil, nil)
		if err != nil {
			panic(err)
		}
		h.cl[instance.AuthURL] = cli

	}
	return h.cl[instance.AuthURL], nil

}

func (h HandlerPluginImpl) AddFlavour(vimInstance *catalogue.VIMInstance, deploymentFlavour *catalogue.DeploymentFlavour) (*catalogue.DeploymentFlavour, error) {
	return deploymentFlavour, nil
}
func (h HandlerPluginImpl) AddImage(vimInstance *catalogue.VIMInstance, image *catalogue.NFVImage, imageFile []byte) (*catalogue.NFVImage, error) {
	return image, nil
}
func (h HandlerPluginImpl) AddImageFromURL(vimInstance *catalogue.VIMInstance, image *catalogue.NFVImage, imageURL string) (*catalogue.NFVImage, error) {
	return image, nil
}
func (h HandlerPluginImpl) CopyImage(vimInstance *catalogue.VIMInstance, image *catalogue.NFVImage, imageFile []byte) (*catalogue.NFVImage, error) {
	return image, nil
}
func (h HandlerPluginImpl) CreateNetwork(vimInstance *catalogue.VIMInstance, network *catalogue.Network) (*catalogue.Network, error) {
	cl, err := h.getClient(vimInstance)
	if err != nil {
		h.logger.Errorf("Error getting the client: %v", err)
		return nil, err
	}
	ipamConfig := make([]network2.IPAMConfig, 1)
	ipamConfig[0].Subnet = network.Subnets[0].CIDR
	resp,err := cl.NetworkCreate(h.ctx, network.Name, types.NetworkCreate{
		IPAM: &network2.IPAM{
			Config: ipamConfig,
		},
	})
	if err != nil{
		h.logger.Errorf("Error creating network: %v", err)
		return nil, err
	}
	net, err := GetNetworkCreate(network.Subnets[0].CIDR, network.Name, resp)
	if err != nil {
		return nil, err
	}
	return &net, nil
}

func (h HandlerPluginImpl) CreateSubnet(vimInstance *catalogue.VIMInstance, createdNetwork *catalogue.Network, subnet *catalogue.Subnet) (*catalogue.Subnet, error) {
	return subnet, nil
}
func (h HandlerPluginImpl) DeleteFlavour(vimInstance *catalogue.VIMInstance, extID string) (bool, error) {
	return true, nil
}
func (h HandlerPluginImpl) DeleteImage(vimInstance *catalogue.VIMInstance, image *catalogue.NFVImage) (bool, error) {
	return true, nil
}
func (h HandlerPluginImpl) DeleteNetwork(vimInstance *catalogue.VIMInstance, extID string) (bool, error) {
	return true, nil
}
func (h HandlerPluginImpl) DeleteServerByIDAndWait(vimInstance *catalogue.VIMInstance, id string) error {
	return nil
}
func (h HandlerPluginImpl) DeleteSubnet(vimInstance *catalogue.VIMInstance, existingSubnetExtID string) (bool, error) {
	return true, nil
}
func (h HandlerPluginImpl) LaunchInstance(vimInstance *catalogue.VIMInstance, name, image, Flavour, keypair string, network []*catalogue.VNFDConnectionPoint, secGroup []string, userData string) (*catalogue.Server, error) {
	srv := &catalogue.Server{}
	return srv, nil
}
func (h HandlerPluginImpl) LaunchInstanceAndWait(vimInstance *catalogue.VIMInstance, hostname, image, extID, keyPair string, network []*catalogue.VNFDConnectionPoint, securityGroups []string, s string) (*catalogue.Server, error) {
	srv := &catalogue.Server{}
	return srv, nil
}
func (h HandlerPluginImpl) LaunchInstanceAndWaitWithIPs(vimInstance *catalogue.VIMInstance, hostname, image, extID, keyPair string, network []*catalogue.VNFDConnectionPoint, securityGroups []string, s string, floatingIps map[string]string, keys []*catalogue.Key) (*catalogue.Server, error) {
	srv := &catalogue.Server{}
	return srv, nil
}
func (h HandlerPluginImpl) ListFlavours(vimInstance *catalogue.VIMInstance) ([]*catalogue.DeploymentFlavour, error) {
	_, err := h.getClient(vimInstance)
	if err != nil {
		h.logger.Errorf("Error getting client: %v", err)
		return nil, err
	}

	res := make([]*catalogue.DeploymentFlavour, 1)

	res[0] = &catalogue.DeploymentFlavour{
		ExtID:      "12345",
		FlavourKey: "m1.small",
		Disk:       0,
		RAM:        0,
		VCPUs:      0,
	}
	return res, nil
}
func (h HandlerPluginImpl) ListImages(vimInstance *catalogue.VIMInstance) ([]*catalogue.NFVImage, error) {

	cl, err := h.getClient(vimInstance)
	if err != nil {
		h.logger.Errorf("Error getting client: %v", err)
		return nil, err
	}
	opt := types.ImageListOptions{}
	images, err := cl.ImageList(h.ctx, opt)
	if err != nil {
		h.logger.Errorf("Error listing images: %v", err)
		return nil, err
	}

	res := make([]*catalogue.NFVImage, len(images))

	for index, img := range images {
		nfvImg, err := GetImage(img)
		if err != nil {
			h.logger.Errorf("Error translating image: %v", err)
			return nil, err
		}
		res[index] = nfvImg
	}
	return res, nil
}
func (h HandlerPluginImpl) ListNetworks(vimInstance *catalogue.VIMInstance) ([]*catalogue.Network, error) {
	cl, err := h.getClient(vimInstance)
	if err != nil {
		h.logger.Errorf("Error getting client: %v", err)
		return nil, err
	}
	opt := types.NetworkListOptions{}
	networksDock, err := cl.NetworkList(h.ctx, opt)
	if err != nil {
		h.logger.Errorf("Error listing networks: %v", err)
		return nil, err
	}

	res := make([]*catalogue.Network, len(networksDock))

	for index, netDock := range networksDock {
		nfvImg, err := GetNetwork(netDock)
		if err != nil {
			h.logger.Errorf("Error translating image: %v", err)
			return nil, err
		}
		res[index] = nfvImg
	}
	return res, nil
}

func (h HandlerPluginImpl) ListServer(vimInstance *catalogue.VIMInstance) ([]*catalogue.Server, error) {
	cl, err := h.getClient(vimInstance)
	if err != nil {
		h.logger.Errorf("Error getting client: %v", err)
		return nil, err
	}
	opt := types.ContainerListOptions{}
	containers, err := cl.ContainerList(h.ctx, opt)
	if err != nil {
		h.logger.Errorf("Error listing networks: %v", err)
		return nil, err
	}

	res := make([]*catalogue.Server, len(containers))

	for index, container := range containers {
		img, err := h.getImageById(container.Image, cl)
		if err != nil {
			h.logger.Errorf("Error while retrieving the image by id")
			return nil, err
		}
		server, err := GetContainer(container, img)
		if err != nil {
			h.logger.Errorf("Error translating image: %v", err)
			return nil, err
		}
		res[index] = server
	}
	return res, nil
}
func (h HandlerPluginImpl) getImageById(i string, cl *client.Client) (*catalogue.NFVImage, error) {
	//filter := filters.NewArgs()
	//filter.Add("id", i)
	//f := filters.Args{}
	//opt := types.ImageListOptions{
	//	Filters:f,
	//}
	images, err := cl.ImageList(h.ctx, types.ImageListOptions{})
	if err != nil {
		h.logger.Errorf("Error listing images: %v", err)
		return nil, err
	}
	for _, img := range images {
		if strings.HasPrefix(img.ID, i) || img.ID == i {
			img, err := GetImage(img)
			return img, err
		}
	}
	return nil, errors.New(fmt.Sprintf("Image with id %s not found", i))
}

func (h HandlerPluginImpl) NetworkByID(vimInstance *catalogue.VIMInstance, id string) (*catalogue.Network, error) {
	return nil, nil
}
func (h HandlerPluginImpl) Quota(vimInstance *catalogue.VIMInstance) (*catalogue.Quota, error) {
	return &catalogue.Quota{
		RAM:         100000,
		Cores:       100000,
		FloatingIPs: 100000,
		KeyPairs:    100000,
		Instances:   100000,
	}, nil
}
func (h HandlerPluginImpl) SubnetsExtIDs(vimInstance *catalogue.VIMInstance, networkExtID string) ([]string, error) {
	return nil, nil
}
func (h HandlerPluginImpl) Type(vimInstance *catalogue.VIMInstance) (string, error) {
	return "docker", nil
}
func (h HandlerPluginImpl) UpdateFlavour(vimInstance *catalogue.VIMInstance, deploymentFlavour *catalogue.DeploymentFlavour) (*catalogue.DeploymentFlavour, error) {
	return deploymentFlavour, nil
}
func (h HandlerPluginImpl) UpdateImage(vimInstance *catalogue.VIMInstance, image *catalogue.NFVImage) (*catalogue.NFVImage, error) {
	return image, nil
}
func (h HandlerPluginImpl) UpdateNetwork(vimInstance *catalogue.VIMInstance, network *catalogue.Network) (*catalogue.Network, error) {
	return network, nil
}
func (h HandlerPluginImpl) UpdateSubnet(vimInstance *catalogue.VIMInstance, createdNetwork *catalogue.Network, subnet *catalogue.Subnet) (*catalogue.Subnet, error) {
	return subnet, nil
}
