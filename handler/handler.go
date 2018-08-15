package handler

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	_ "net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"io/ioutil"
	"math/rand"

	"docker.io/go-docker"
	"docker.io/go-docker/api"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	dockerNetwork "docker.io/go-docker/api/types/network"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/op/go-logging"
	"github.com/openbaton/go-openbaton/catalogue"
	"github.com/openbaton/go-openbaton/pluginsdk"
	"github.com/openbaton/go-openbaton/sdk"
)

var dockerSecDir = "docker_sec"

type PluginImpl struct {
	Logger        *logging.Logger
	ctx           context.Context
	cl            map[string]*docker.Client
	Swarm         bool
	Tsl           bool
	CertDirectory string
}

func NewHandlerPlugin(swarm bool) *PluginImpl {
	return &PluginImpl{
		Logger: sdk.GetLogger("HandlerPlugin", "DEBUG"),
		Swarm:  swarm,
	}
}

func (h *PluginImpl) getClient(instance *catalogue.DockerVimInstance) (*docker.Client, error) {
	var dir string
	if instance.Ca != "" {
		var err error
		if !exists(dockerSecDir) {
			dir, err = ioutil.TempDir(dockerSecDir, "")
		}
		if err != nil {
			h.Logger.Errorf("Error creating temp dir")
			return nil, err
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s/ca.pem", dir), []byte(instance.Ca), os.ModePerm)
		err = ioutil.WriteFile(fmt.Sprintf("%s/cert.pem", dir), []byte(instance.Ca), os.ModePerm)
		err = ioutil.WriteFile(fmt.Sprintf("%s/key.pem", dir), []byte(instance.Ca), os.ModePerm)
	}

	if h.ctx == nil {
		h.ctx = context.Background()
	}

	if h.cl == nil {
		h.cl = make(map[string]*docker.Client)
	}

	if _, ok := h.cl[instance.AuthURL]; !ok {
		var cli *docker.Client
		var err error
		if strings.HasPrefix(instance.AuthURL, "unix:") {
			cli, err = docker.NewClient(instance.AuthURL, api.DefaultVersion, nil, nil)
		} else {
			var tlsc *tls.Config
			if h.Tsl {

				options := tlsconfig.Options{
					CAFile:             filepath.Join(dir, "ca.pem"),
					CertFile:           filepath.Join(dir, "cert.pem"),
					KeyFile:            filepath.Join(dir, "key.pem"),
					InsecureSkipVerify: false,
				}
				tlsc, err = tlsconfig.Client(options)
				if err != nil {
					return nil, err
				}
			}
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: tlsc,
				},
				CheckRedirect: docker.CheckRedirect,
			}
			cli, err = docker.NewClient(instance.AuthURL, api.DefaultVersion, httpClient, nil)
			os.RemoveAll(dir)
		}
		if err != nil {
			panic(err)
		}
		h.cl[instance.AuthURL] = cli
	}

	return h.cl[instance.AuthURL], nil

}

func (h PluginImpl) AddFlavour(vimInstance interface{}, deploymentFlavour *catalogue.DeploymentFlavour) (*catalogue.DeploymentFlavour, error) {
	return deploymentFlavour, nil
}

func (h PluginImpl) AddImage(vimInstance interface{}, image catalogue.BaseImageInt, imageFile []byte) (catalogue.BaseImageInt, error) {
	return image, nil
}

func (h PluginImpl) AddImageFromURL(vimInstance interface{}, image catalogue.BaseImageInt, imageURL string) (catalogue.BaseImageInt, error) {
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting Docker Vim Instance: %v", err)
		return nil, err
	}
	dockerImage, err := getDockerImage(image)
	if err != nil {
		h.Logger.Errorf("Error getting Docker image: %v", err)
		return nil, err
	}
	cl, err := h.getClient(dockerVimInstance)
	if err != nil {
		h.Logger.Errorf("Error while getting client: %v", err)
		return nil, err
	}
	h.Logger.Noticef("Trying to pull image: %v", imageURL)
	out, err := cl.ImagePull(h.ctx, imageURL, types.ImagePullOptions{
		All: false,
	})

	if err != nil {
		h.Logger.Errorf("Not able to pull image %s: %v", imageURL, err)
		return nil, err
	}

	io.Copy(os.Stdout, out)
	defer out.Close()

	img, err := getImagesByName(cl, h.ctx, imageURL)

	h.Logger.Debugf("New Tags are: %v", img[0].RepoTags)
	if len(img) == 1 {
		var extId string
		if strings.Contains(img[0].ID, ":") {
			extId = strings.Split(img[0].ID, ":")[1]
		} else {
			extId = img[0].ID
		}
		dockerImage.ExtID = extId
		dockerImage.Tags = img[0].RepoTags
		dockerImage.Created = catalogue.NewDateWithTime(time.Now())
	}

	return dockerImage, err
}

func getImagesByName(cl *docker.Client, ctx context.Context, imageName string) ([]types.ImageSummary, error) {
	//var args filters.Args
	//args = filters.NewArgs(filters.KeyValuePair{
	//	Key:   "repotag",
	//	Value: imageName,
	//})
	imgs, err := cl.ImageList(ctx, types.ImageListOptions{})
	res := make([]types.ImageSummary, 0)
	if err != nil {
		return nil, err
	}
	for _, img := range imgs {
		if stringInSlice(imageName, img.RepoTags) {
			res = append(res, img)
		}
	}
	if len(res) == 0 {
		return nil, errors.New(fmt.Sprintf("Image with name %s not found", imageName))
	}
	return res, nil
}

func (h PluginImpl) CopyImage(vimInstance interface{}, image catalogue.BaseImageInt, imageFile []byte) (catalogue.BaseImageInt, error) {
	return image, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func (h PluginImpl) CreateNetwork(vimInstance interface{}, network catalogue.BaseNetworkInt) (catalogue.BaseNetworkInt, error) {
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting Docker Vim Instance: %v", err)
		return nil, err
	}
	dockerNet, err := getDockerNet(network)
	if err != nil {
		h.Logger.Errorf("Error getting the Docker Network: %v", err)
		return nil, err
	}
	cl, err := h.getClient(dockerVimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting the client: %v", err)
		return nil, err
	}
	var driver string
	if h.Swarm {
		driver = "overlay"
	} else {
		driver = "bridge"
	}

	var ipam *dockerNetwork.IPAM
	if dockerNet.Subnet != "" {
		ipamConfig := make([]dockerNetwork.IPAMConfig, 1)
		ipamConfig[0].Subnet = dockerNet.Subnet
		ip, _, err := net.ParseCIDR(dockerNet.Subnet)
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		inc(ip)
		ipamConfig[0].Gateway = ip.String()
		ipam = &dockerNetwork.IPAM{
			Config: ipamConfig,
		}
	}
	h.Logger.Debugf("Received DockerNetwork %+v", dockerNet)
	dockerNet.Name = fmt.Sprintf("%s_%d", dockerNet.Name, 9999-rand.Intn(9000))
	for ok, err := existsNetwork(cl, h.ctx, dockerNet.Name); ok; ok, err = existsNetwork(cl, h.ctx, dockerNet.Name) {

		if err != nil {
			h.Logger.Errorf("Not able to list network with name %s: %v", dockerNet.Name, err)
			return nil, err
		}
		dockerNet.Name = fmt.Sprintf("%s_%d", strings.Split(dockerNet.Name, "_")[0], 9999-rand.Intn(9000))
		ok, err = existsNetwork(cl, h.ctx, dockerNet.Name)
	}

	netCreateOpt := types.NetworkCreate{
		IPAM:   ipam,
		Driver: driver,
	}
	h.Logger.Debugf("Creating network [%s] with config %v", dockerNet.Name, netCreateOpt)
	resp, err := cl.NetworkCreate(h.ctx, dockerNet.Name, netCreateOpt)
	if err != nil {
		h.Logger.Errorf("Error creating network: %v", err)
		return nil, err
	}
	dockNet, err := cl.NetworkInspect(h.ctx, resp.ID, types.NetworkInspectOptions{})
	if err != nil {
		h.Logger.Errorf("Error inspecting network: %v", err)
		return nil, err
	}
	obNet, err := GetNetwork(dockNet)
	h.Logger.Infof("Created ob network [%s] with ext id [%s]", obNet.Name, obNet.ExtID)
	if err != nil {
		return nil, err
	}
	return obNet, nil
}
func existsNetwork(cl *docker.Client, ctx context.Context, name string) (bool, error) {
	keyValuePair := filters.NewArgs(filters.Arg("name", name))
	nets, err := cl.NetworkList(ctx, types.NetworkListOptions{
		Filters: keyValuePair,
	})
	if err != nil {
		return false, err
	}
	return len(nets) > 0, nil
}

func (h PluginImpl) CreateSubnet(vimInstance interface{}, createdNetwork catalogue.BaseNetworkInt, subnet *catalogue.Subnet) (*catalogue.Subnet, error) {
	return subnet, nil
}
func (h PluginImpl) DeleteFlavour(vimInstance interface{}, extID string) (bool, error) {
	return true, nil
}
func (h PluginImpl) DeleteImage(vimInstance interface{}, image catalogue.BaseImageInt) (bool, error) {
	return true, nil
}
func (h PluginImpl) DeleteNetwork(vimInstance interface{}, extID string) (bool, error) {
	h.Logger.Debugf("Deleting network [%s]", extID)
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting docker vim instance: %v", err)
		return false, err
	}
	cl, err := h.getClient(dockerVimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting client: %v", err)
		return false, err
	}
	err = cl.NetworkRemove(h.ctx, extID)
	if err != nil {
		h.Logger.Errorf("Error Deleting network: %v", err)
		return false, err
	}
	h.Logger.Infof("Deleted network [%s]", extID)
	return true, nil
}
func (h PluginImpl) DeleteServerByIDAndWait(vimInstance interface{}, id string) error {
	return nil
}
func (h PluginImpl) DeleteSubnet(vimInstance interface{}, existingSubnetExtID string) (bool, error) {
	return true, nil
}
func (h PluginImpl) LaunchInstance(vimInstance interface{}, name, image, Flavour, keypair string, network []*catalogue.VNFDConnectionPoint, secGroup []string, userData string) (*catalogue.Server, error) {
	srv := &catalogue.Server{}
	return srv, nil
}
func (h PluginImpl) LaunchInstanceAndWait(vimInstance interface{}, hostname, image, flavorKey, keyPair string, network []*catalogue.VNFDConnectionPoint, securityGroups []string, userdata string) (*catalogue.Server, error) {
	if userdata != "" {
		h.Logger.Warning("User-data is IGNORED, why did you pass it?!")
	}
	srv := &catalogue.Server{}
	return srv, nil
}
func (h PluginImpl) LaunchInstanceAndWaitWithIPs(vimInstance interface{}, hostname, image, extID, keyPair string, network []*catalogue.VNFDConnectionPoint, securityGroups []string, userdata string, floatingIps map[string]string, keys []*catalogue.Key) (*catalogue.Server, error) {

	return h.LaunchInstanceAndWait(vimInstance, hostname, image, extID, keyPair, network, securityGroups, userdata)
}
func (h PluginImpl) ListFlavours(vimInstance interface{}) ([]*catalogue.DeploymentFlavour, error) {
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting Docker Vim Instance: %v", err)
		return nil, err
	}
	_, err = h.getClient(dockerVimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting client: %v", err)
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
func (h PluginImpl) ListImages(vimInstance interface{}) (catalogue.BaseImageInt, error) {
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting Docker Vim Instance: %v", err)
		return nil, err
	}
	cl, err := h.getClient(dockerVimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting client: %v", err)
		return nil, err
	}
	opt := types.ImageListOptions{}
	images, err := cl.ImageList(h.ctx, opt)
	if err != nil {
		h.Logger.Errorf("Error listing images: %v", err)
		return nil, err
	}

	res := make([]*catalogue.DockerImage, len(images))

	for index, img := range images {
		nfvImg, err := GetImage(img)
		if err != nil {
			h.Logger.Errorf("Error translating image: %v", err)
			return nil, err
		}
		res[index] = nfvImg
	}
	h.Logger.Infof("Listed %d images", len(res))
	return res, nil
}
func (h PluginImpl) ListNetworks(vimInstance interface{}) (catalogue.BaseNetworkInt, error) {
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting Docker Vim Instance: %v", err)
		return nil, err
	}
	cl, err := h.getClient(dockerVimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting client: %v", err)
		return nil, err
	}
	opt := types.NetworkListOptions{}
	networksDock, err := cl.NetworkList(h.ctx, opt)
	if err != nil {
		h.Logger.Errorf("Error listing networks: %v", err)
		return nil, err
	}

	res := make([]*catalogue.DockerNetwork, len(networksDock))

	for index, netDock := range networksDock {
		obNet, err := GetNetwork(netDock)
		if err != nil {
			h.Logger.Errorf("Error translating image: %v", err)
			return nil, err
		}
		res[index] = obNet
	}
	h.Logger.Infof("Listed %d networks", len(res))
	h.Logger.Debugf("Networks are %s", func() []string {
		v := make([]string, len(res))
		for i, n := range res {
			v[i] = n.Name
		}
		return v
	}())
	return res, nil
}

func (h PluginImpl) Refresh(vimInstance interface{}) (interface{}, error) {
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting Docker Vim Instance: %v", err)
		return nil, err
	}
	//Images
	imgs, err := h.ListImages(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error listing images: %v", err)
		return nil, err
	}
	imageLen := len(imgs.([]*catalogue.DockerImage))
	dockerVimInstance.Images = make([]catalogue.DockerImage, imageLen)
	for i := 0; i < imageLen; i++ {
		dockerVimInstance.Images[i] = *(imgs.([]*catalogue.DockerImage)[i])
	}
	//Networks
	nets, err := h.ListNetworks(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error listing networks: %v", err)
		return nil, err
	}
	netLen := len(nets.([]*catalogue.DockerNetwork))
	dockerVimInstance.Networks = make([]catalogue.DockerNetwork, netLen)
	for i := 0; i < netLen; i++ {
		dockerVimInstance.Networks[i] = *(nets.([]*catalogue.DockerNetwork)[i])
	}

	return dockerVimInstance, nil
}

func (h PluginImpl) ListServer(vimInstance interface{}) ([]*catalogue.Server, error) {
	dockerVimInstance, err := pluginsdk.GetDockerVimInstance(vimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting Docker Vim Instance: %v", err)
		return nil, err
	}
	cl, err := h.getClient(dockerVimInstance)
	if err != nil {
		h.Logger.Errorf("Error getting client: %v", err)
		return nil, err
	}
	opt := types.ContainerListOptions{}
	containers, err := cl.ContainerList(h.ctx, opt)
	if err != nil {
		h.Logger.Errorf("Error listing networks: %v", err)
		return nil, err
	}

	res := make([]*catalogue.Server, 0)

	for _, container := range containers {
		img, err := h.getImageById(container.Image, cl)
		dimg, err := getDockerImage(img)
		var server *catalogue.Server
		if err != nil {
			h.Logger.Errorf("Error while retrieving the image by id")
			imageSummary, _, err := cl.ImageInspectWithRaw(h.ctx, dimg.ID)
			if err != nil {
				h.Logger.Errorf("Error inspecting image: %v", err)
				return nil, err
			}
			server, err = GetContainerWithImgName(container, imageSummary)
			// return nil, err
		}
		server, err = GetContainer(container, dimg)
		if err != nil {
			h.Logger.Errorf("Error translating image: %v", err)
			return nil, err
		}
		res = append(res, server)
	}
	return res, nil
}

func (h PluginImpl) getImageById(i string, cl *docker.Client) (catalogue.BaseImageInt, error) {
	//filter := filters.NewArgs()
	//filter.Add("id", i)
	//f := filters.Args{}
	//opt := types.ImageListOptions{
	//	Filters:f,
	//}
	images, err := cl.ImageList(h.ctx, types.ImageListOptions{})
	if err != nil {
		h.Logger.Errorf("Error listing images: %v", err)
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

func (h PluginImpl) NetworkByID(vimInstance interface{}, id string) (catalogue.BaseNetworkInt, error) {
	return nil, nil
}
func (h PluginImpl) Quota(vimInstance interface{}) (*catalogue.Quota, error) {
	return &catalogue.Quota{
		RAM:         100000,
		Cores:       100000,
		FloatingIPs: 100000,
		KeyPairs:    100000,
		Instances:   100000,
	}, nil
}
func (h PluginImpl) SubnetsExtIDs(vimInstance interface{}, networkExtID string) ([]string, error) {
	return nil, nil
}
func (h PluginImpl) Type(vimInstance interface{}) (string, error) {
	return "docker", nil
}
func (h PluginImpl) UpdateFlavour(vimInstance interface{}, deploymentFlavour *catalogue.DeploymentFlavour) (*catalogue.DeploymentFlavour, error) {
	return deploymentFlavour, nil
}
func (h PluginImpl) UpdateImage(vimInstance interface{}, image catalogue.BaseImageInt) (catalogue.BaseImageInt, error) {
	return image, nil
}
func (h PluginImpl) UpdateNetwork(vimInstance interface{}, network catalogue.BaseNetworkInt) (catalogue.BaseNetworkInt, error) {
	return network, nil
}
func (h PluginImpl) UpdateSubnet(vimInstance interface{}, createdNetwork catalogue.BaseNetworkInt, subnet *catalogue.Subnet) (*catalogue.Subnet, error) {
	return subnet, nil
}
func (h PluginImpl) RebuildServer(vimInstance interface{}, serverId string, imageId string) (*catalogue.Server, error) {
	srv := &catalogue.Server{}
	return srv, nil
}
