/* Copyright 2020 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package orchestrator

import (
	"fmt"
	"io"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	log "github.com/golang/glog"
	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/moby/moby/client"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

// Client used to interface with Docker.
type Client interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
}

var cli Client

var newClient = func() {
	if cli == nil {
		var err error
		cli, err = client.NewEnvClient()
		if err != nil {
			log.Exitf("couldn't create docker client: %v", err)
		}
	}
}

// InitContainers will check if the containers are running and run them if not.
func InitContainers(names []string) error {
	newClient()
	build := viper.GetString("docker.build")
	if err := pullImage(build); err != nil {
		return err
	}
	runtime := viper.GetString("docker.runtime")
	if err := pullImage(runtime); err != nil {
		return err
	}

	tests := config.GetTests()
	if len(names) == 0 {
		names = make([]string, len(tests))
		i := 0
		for name := range tests {
			names[i] = name
			i++
		}
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}
	for _, c := range containers {
		for _, name := range c.Names {
			for i, testName := range names {
				if name == testName {
					if c.Status != "running" {
						if err := cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{}); err != nil {
							return err
						}
					}
					copy(names[i:], names[i+1:])
					names[len(names)-1] = ""
					names = names[:len(names)-1]
				}
			}
		}
	}
	for _, name := range names {
		if err := createContainer(name); err != nil {
			return err
		}
	}
	return nil
}

// createContainer will build the image and run the container.
func createContainer(name string) error {
	found, err := findImage(name)
	if err != nil {
		return nil
	}
	if !found {
		_, err := cli.ImageBuild(
			context.Background(),
			nil,
			types.ImageBuildOptions{
				Dockerfile: path.Join(viper.GetString("docker.files"), fmt.Sprintf("%s.Dockerfile", name)),
				Tags:       []string{name},
			},
		)
		if err != nil {
			return err
		}
	}
	c, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{Image: name},
		&container.HostConfig{},
		&network.NetworkingConfig{},
		name,
	)
	if err != nil {
		return err
	}
	if err := cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

func pullImage(name string) error {
	found, err := findImage(name)
	if err != nil {
		return nil
	}
	if !found {
		closer, err := cli.ImagePull(context.Background(), name, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		closer.Close()
	}
	return nil
}

func findImage(name string) (bool, error) {
	list, err := cli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return false, err
	}
	found := false
imageCheck:
	for _, img := range list {
		for _, tag := range img.RepoTags {
			if tag == name {
				found = true
				break imageCheck
			}
		}
	}
	return found, nil
}

// RunContainer runs an executable in a docker conatainer.
var RunContainer = func(name, args string) (out string, code int, err error) {
	return "", 0, nil
}
