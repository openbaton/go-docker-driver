/*
 * Copyright (c) 2016 Fraunhofer FOKUS
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package org.openbaton.docker;

import com.github.dockerjava.api.DockerClient;
import com.github.dockerjava.api.command.CreateContainerResponse;
import com.github.dockerjava.api.model.*;
import com.github.dockerjava.core.DefaultDockerClientConfig;
import com.github.dockerjava.core.DockerClientBuilder;
import com.github.dockerjava.core.DockerClientConfig;
import org.openbaton.catalogue.mano.common.DeploymentFlavour;
import org.openbaton.catalogue.nfvo.*;
import org.openbaton.catalogue.nfvo.Network;
import org.openbaton.catalogue.security.Key;
import org.openbaton.exceptions.VimDriverException;
import org.openbaton.vim.drivers.interfaces.VimDriver;
import org.openbaton.plugin.PluginStarter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.util.*;
import java.util.concurrent.TimeoutException;

/**
 * Created by gca on 24/08/16.
 */
public class DockerVim extends VimDriver {
  private static final Logger log = LoggerFactory.getLogger(DockerVim.class);

  public static void main(String[] args)
      throws NoSuchMethodException, IOException, InstantiationException, TimeoutException,
          IllegalAccessException, InvocationTargetException {
    if (args.length <= 1)
      PluginStarter.registerPlugin(DockerVim.class, "docker", "localhost", 5672, 3);
    else
      PluginStarter.registerPlugin(
          DockerVim.class,
          args[0],
          args[1],
          Integer.parseInt(args[2]),
          Integer.parseInt(args[3]),
          args[4],
          args[5]);

    /*    VimInstance vimInstance = new VimInstance();
    vimInstance.setAuthUrl("tcp://localhost:2375");
    DockerVim dockerVim = new DockerVim();
    try {
      dockerVim.listImages(vimInstance);
    } catch (VimDriverException e) {
      e.printStackTrace();
    }
    try {
      dockerVim.listNetworks(vimInstance);
    } catch (VimDriverException e) {
      e.printStackTrace();
    }*/
  }

  private DockerClient createClient(String endpoint) {
    DockerClientConfig config =
        DefaultDockerClientConfig.createDefaultConfigBuilder().withDockerHost(endpoint).build();
    return DockerClientBuilder.getInstance(config).build();
  }

  @Override
  public Server launchInstance(
      VimInstance vimInstance,
      String name,
      String image,
      String flavor,
      String keypair,
      Set<String> network,
      Set<String> secGroup,
      String userData)
      throws VimDriverException {
    this.createClient(vimInstance.getAuthUrl());

    return null;
  }

  @Override
  public List<NFVImage> listImages(VimInstance vimInstance) throws VimDriverException {
    List<NFVImage> images = new ArrayList<>();
    DockerClient docker = this.createClient(vimInstance.getAuthUrl());
    List<Image> dockerImages = docker.listImagesCmd().withShowAll(true).exec();
    for (Image image : dockerImages) {
      NFVImage nfvImage = new NFVImage();
      nfvImage.setName(image.getId());
      nfvImage.setContainerFormat("docker");
      nfvImage.setCreated(new Date(image.getCreated()));
      nfvImage.setIsPublic(true);
      log.debug("Found a docker image, transformed into a NFV Image " + nfvImage);
      images.add(nfvImage);
    }

    return images;
  }

  @Override
  public List<Server> listServer(VimInstance vimInstance) throws VimDriverException {
    return null;
  }

  @Override
  public List<Network> listNetworks(VimInstance vimInstance) throws VimDriverException {
    List<Network> networks = new ArrayList<>();
    DockerClient docker = this.createClient(vimInstance.getAuthUrl());
    List<com.github.dockerjava.api.model.Network> dockerNetworks = docker.listNetworksCmd().exec();
    for (com.github.dockerjava.api.model.Network dockerNetwork : dockerNetworks) {
      Network net = new Network();
      net.setName(dockerNetwork.getName());
      Set<Subnet> subnets = new HashSet<>();
      Subnet subnet = new Subnet();
      subnet.setName(dockerNetwork.getName());
      subnets.add(subnet);
      net.setSubnets(subnets);
      log.debug("Found a docker network, transformed into a NFV network " + net);

      networks.add(net);
    }
    return networks;
  }

  @Override
  public List<DeploymentFlavour> listFlavors(VimInstance vimInstance) throws VimDriverException {
    // flavors don't exist in docker - invalid method
    return new ArrayList<>();
  }

  @Override
  public Server launchInstanceAndWait(
      VimInstance vimInstance,
      String hostname,
      String image,
      String extId,
      String keyPair,
      Set<String> networks,
      Set<String> securityGroups,
      String s,
      Map<String, String> floatingIps,
      Set<Key> keys)
      throws VimDriverException {
    return null;
  }

  @Override
  public Server launchInstanceAndWait(
      VimInstance vimInstance,
      String hostname,
      String image,
      String extId,
      String keyPair,
      Set<String> networks,
      Set<String> securityGroups,
      String s)
      throws VimDriverException {
    return null;
  }

  @Override
  public void deleteServerByIdAndWait(VimInstance vimInstance, String id)
      throws VimDriverException {}

  @Override
  public Network createNetwork(VimInstance vimInstance, Network network) throws VimDriverException {
    return null;
  }

  @Override
  public DeploymentFlavour addFlavor(VimInstance vimInstance, DeploymentFlavour deploymentFlavour)
      throws VimDriverException {
    return null;
  }

  @Override
  public NFVImage addImage(VimInstance vimInstance, NFVImage image, byte[] imageFile)
      throws VimDriverException {
    return null;
  }

  @Override
  public NFVImage addImage(VimInstance vimInstance, NFVImage image, String image_url)
      throws VimDriverException {
    return null;
  }

  @Override
  public NFVImage updateImage(VimInstance vimInstance, NFVImage image) throws VimDriverException {
    return null;
  }

  @Override
  public NFVImage copyImage(VimInstance vimInstance, NFVImage image, byte[] imageFile)
      throws VimDriverException {
    return null;
  }

  @Override
  public boolean deleteImage(VimInstance vimInstance, NFVImage image) throws VimDriverException {
    return false;
  }

  @Override
  public DeploymentFlavour updateFlavor(
      VimInstance vimInstance, DeploymentFlavour deploymentFlavour) throws VimDriverException {
    return null;
  }

  @Override
  public boolean deleteFlavor(VimInstance vimInstance, String extId) throws VimDriverException {
    return false;
  }

  @Override
  public Subnet createSubnet(VimInstance vimInstance, Network createdNetwork, Subnet subnet)
      throws VimDriverException {
    return null;
  }

  @Override
  public Network updateNetwork(VimInstance vimInstance, Network network) throws VimDriverException {
    return null;
  }

  @Override
  public Subnet updateSubnet(VimInstance vimInstance, Network updatedNetwork, Subnet subnet)
      throws VimDriverException {
    return null;
  }

  @Override
  public List<String> getSubnetsExtIds(VimInstance vimInstance, String network_extId)
      throws VimDriverException {
    return null;
  }

  @Override
  public boolean deleteSubnet(VimInstance vimInstance, String existingSubnetExtId)
      throws VimDriverException {
    return false;
  }

  @Override
  public boolean deleteNetwork(VimInstance vimInstance, String extId) throws VimDriverException {
    return false;
  }

  @Override
  public Network getNetworkById(VimInstance vimInstance, String id) throws VimDriverException {
    return null;
  }

  @Override
  public Quota getQuota(VimInstance vimInstance) throws VimDriverException {
    return null;
  }

  @Override
  public String getType(VimInstance vimInstance) throws VimDriverException {
    return null;
  }
}
