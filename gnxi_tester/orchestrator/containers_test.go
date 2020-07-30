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
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

type clientCounter struct {
	CountContainerList,
	CountContainerStart,
	CountImageList,
	CountImageBuild,
	CountImagePull,
	CountContainerCreate int
}

type mockClient struct {
	Client
	clientCounter
	images     []types.ImageSummary
	containers []types.Container
}

type mockReader int

func (r mockReader) Read(p []byte) (n int, err error) {
	return
}

func (r mockReader) Close() error {
	return nil
}

func (c *mockClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	c.CountContainerList++
	return c.containers, nil
}

func (c *mockClient) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	c.CountContainerStart++
	return nil
}

func (c *mockClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	c.CountImageList++
	return c.images, nil
}

func (c *mockClient) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	c.CountImageBuild++
	return types.ImageBuildResponse{}, nil
}

func (c *mockClient) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	c.CountImagePull++
	return mockReader(0), nil
}

func (c *mockClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	c.CountContainerCreate++
	return container.ContainerCreateCreatedBody{}, nil
}

func TestInitContainer(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		err   error
		buildImg,
		runtimeImg string
		images     []types.ImageSummary
		containers []types.Container
		counter    clientCounter
	}{
		{
			"container already running",
			[]string{"name"},
			nil,
			"build",
			"runtime",
			[]types.ImageSummary{
				{RepoTags: []string{"name"}},
				{RepoTags: []string{"build"}},
				{RepoTags: []string{"runtime"}},
			},
			[]types.Container{{Names: []string{"name"}, Status: "running"}},
			clientCounter{CountImageList: 2, CountContainerList: 1},
		},
		{
			"container not running",
			[]string{"name"},
			nil,
			"build",
			"runtime",
			[]types.ImageSummary{
				{RepoTags: []string{"name"}},
				{RepoTags: []string{"build"}},
				{RepoTags: []string{"runtime"}},
			},
			[]types.Container{{Names: []string{"name"}}},
			clientCounter{CountImageList: 2, CountContainerList: 1, CountContainerStart: 1},
		},
		{
			"pull images and build",
			[]string{"name"},
			nil,
			"build",
			"runtime",
			[]types.ImageSummary{},
			[]types.Container{},
			clientCounter{CountImageList: 3, CountContainerList: 1, CountContainerStart: 1, CountImageBuild: 1, CountImagePull: 2, CountContainerCreate: 1},
		},
		{
			"no name",
			[]string{},
			nil,
			"build",
			"runtime",
			[]types.ImageSummary{},
			[]types.Container{},
			clientCounter{CountImageList: 2, CountContainerList: 1, CountImagePull: 2},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("docker.build", test.buildImg)
			viper.Set("docker.runtime", test.runtimeImg)
			client := &mockClient{images: test.images, containers: test.containers}
			newClient = func() {
				dockerClient = client
			}
			if err := InitContainers(test.names); fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.err) {
				t.Errorf("wanted error(%v), got(%v)", test.err, err)
			}
			if diff := cmp.Diff(test.counter, client.clientCounter); diff != "" {
				t.Errorf("method call counter mismatch (-want +got): %s", diff)
			}
		})
	}
}
