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
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	log "github.com/golang/glog"
	"github.com/google/gnxi/gnxi_tester/config"
	"github.com/mholt/archiver/v3"
	"github.com/moby/moby/client"
	"github.com/moby/moby/pkg/stdcopy"
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
	ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error)
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	ImageRemove(ctx context.Context, imageID string, options types.ImageRemoveOptions) ([]types.ImageDelete, error)
}

var dockerClient Client

var newClient = func() {
	if dockerClient != nil {
		log.Error("docker client exists")
		return
	}
	var err error
	dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Exitf("couldn't create docker client: %v", err)
	}
}

// InitContainers will check if the containers are running and run them if not.
var InitContainers = func(names []string) error {
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
	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}
	for _, c := range containers {
		for _, name := range c.Names {
			if err := checkContainerExists(name, c, &names); err != nil {
				return err
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

func checkContainerExists(containerName string, cont types.Container, names *[]string) error {
	for i, testName := range *names {
		if containerName == "/"+testName {
			if cont.State != "running" {
				if err := dockerClient.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{}); err != nil {
					return err
				}
			}
			copy((*names)[i:], (*names)[i+1:])
			(*names)[len(*names)-1] = ""
			*names = (*names)[:len(*names)-1]
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
		infof("Building image for %s...", name)
		dockerfile := path.Join(viper.GetString("docker.files"), fmt.Sprintf("%s.Dockerfile", name))
		buildContext, err := tarFile(name, dockerfile)
		defer buildContext.Close()
		if err != nil {
			return err
		}
		reader, err := dockerClient.ImageBuild(
			context.Background(),
			buildContext,
			types.ImageBuildOptions{
				Dockerfile: fmt.Sprintf("%s.Dockerfile", name),
				Tags:       []string{fmt.Sprintf("%s:latest", name)},
			},
		)
		if err != nil {
			return err
		}
		io.Copy(stdOut, reader.Body)
		infof("Finished building image for %s", name)
	}
	c, err := dockerClient.ContainerCreate(
		context.Background(),
		&container.Config{Image: fmt.Sprintf("%s:latest", name)},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
			NetworkMode:   container.NetworkMode("host"),
		},
		&network.NetworkingConfig{},
		name,
	)
	if err != nil {
		return err
	}
	if err := dockerClient.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{}); err != nil {
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
		infof("Pulling image %s...", name)
		reader, err := dockerClient.ImagePull(context.Background(), fmt.Sprintf("docker.io/library/%s", name), types.ImagePullOptions{})
		if err != nil {
			return err
		}
		io.Copy(stdOut, reader)
		infof("Finished pulling %s", name)
	}
	return nil
}

func findImage(name string) (bool, error) {
	list, err := dockerClient.ImageList(context.Background(), types.ImageListOptions{All: true})
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

// RunContainer runs an executable in a docker container.
var RunContainer = func(name, args string, device *config.Device, insertFiles []string) (out string, code int, err error) {
	var cont *types.Container
	if cont, err = getContainer(name); err != nil {
		return
	}
	var ca io.ReadCloser
	if ca, err = tarFile("ca", device.Ca); err != nil {
		return
	}
	defer ca.Close()
	if err = dockerClient.CopyToContainer(context.Background(), cont.ID, "/certs", ca, types.CopyToContainerOptions{}); err != nil {
		return
	}
	var key io.ReadCloser
	if key, err = tarFile("key", device.CaKey); err != nil {
		return
	}
	if err = dockerClient.CopyToContainer(context.Background(), cont.ID, "/certs", key, types.CopyToContainerOptions{}); err != nil {
		return
	}
	defer key.Close()
	for _, f := range insertFiles {
		var insert io.ReadCloser
		if insert, err = tarFile(path.Base(f), f); err != nil {
			return
		}
		if err = dockerClient.CopyToContainer(context.Background(), cont.ID, "/tmp", insert, types.CopyToContainerOptions{}); err != nil {
			return
		}
		defer insert.Close()
	}
	command := make([]string, len(args)+1)
	command[0] = name
	if len(args) > 0 {
		for i, arg := range strings.Split(args, " ") {
			command[i+1] = arg
		}
	}
	var id types.IDResponse
	if id, err = dockerClient.ContainerExecCreate(context.Background(), cont.ID, types.ExecConfig{
		Cmd:          command,
		AttachStderr: true,
		AttachStdout: true,
	}); err != nil {
		return
	}
	var (
		resp types.HijackedResponse
		ctx  = context.Background()
		done = make(chan error)
		outBuf,
		errBuf bytes.Buffer
	)
	if resp, err = dockerClient.ContainerExecAttach(ctx, id.ID, types.ExecConfig{}); err != nil {
		err = fmt.Errorf("error starting exec process: %w", err)
		return
	}
	defer resp.Close()
	go func() {
		_, err := stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		done <- err
	}()
	select {
	case err = <-done:
		if err != nil {
			return
		}
	case <-ctx.Done():
		return
	}
	var inspect types.ContainerExecInspect
	if inspect, err = dockerClient.ContainerExecInspect(context.Background(), id.ID); err != nil {
		err = fmt.Errorf("error inspecting exec process: %w", err)
		return
	}
	code = inspect.ExitCode
	out = errBuf.String()
	return
}

func getContainer(name string) (*types.Container, error) {
	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		for _, containerName := range c.Names {
			if containerName == "/"+name {
				return &c, nil
			}
		}
	}
	return nil, fmt.Errorf("couldn't find container %s", name)
}

var tarFile = func(name, filePath string) (io.ReadCloser, error) {
	loc := path.Join("/tmp", fmt.Sprintf("%s.tar.gz", name))
	if _, err := os.Stat(loc); !os.IsNotExist(err) {
		if err = os.Remove(loc); err != nil {
			return nil, err
		}
	}
	if err := archiver.Archive([]string{filePath}, loc); err != nil {
		return nil, err
	}
	f, err := os.Open(loc)
	if err != nil {
		return nil, err
	}
	return f, err
}

// WipeContainers will delete the images and containers generated by the containers
func WipeContainers(names []string) error {
	newClient()
	for _, name := range names {
		cont, err := getContainer(name)
		if err != nil {
			return err
		}
		if err = dockerClient.ContainerRemove(context.Background(), cont.ID, types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
			RemoveLinks:   true,
		}); err != nil {
			return err
		}
		if _, err = dockerClient.ImageRemove(context.Background(), fmt.Sprintf("%s:latest", name), types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		}); err != nil {
			return err
		}
	}
	return nil
}
