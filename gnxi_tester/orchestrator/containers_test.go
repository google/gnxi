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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/google/gnxi/gnxi_tester/config"
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
	CountContainerExecStart,
	CountContainerExecAttach,
	CountContainerExecCreate,
	CountContainerExecInspect,
	CountCopyToContainer,
	CountContainerCreate int
}

type mockClient struct {
	Client
	clientCounter
	images     []types.ImageSummary
	containers []types.Container
	reader     *bufio.Reader
	code       int
}

type mockReader int

func (r mockReader) Read(p []byte) (n int, err error) {
	err = io.EOF
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
	return types.ImageBuildResponse{Body: mockReader(0)}, nil
}

func (c *mockClient) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	c.CountImagePull++
	return mockReader(0), nil
}

func (c *mockClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	c.CountContainerCreate++
	return container.ContainerCreateCreatedBody{}, nil
}

func (c *mockClient) ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error {
	c.CountContainerExecStart++
	return nil
}

func (c *mockClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error) {
	c.CountContainerExecAttach++
	return types.HijackedResponse{
		Reader: c.reader,
		Conn:   &net.UnixConn{},
	}, nil
}

func (c *mockClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	c.CountContainerExecCreate++
	return types.IDResponse{}, nil
}

func (c *mockClient) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	c.CountContainerExecInspect++
	return types.ContainerExecInspect{ExitCode: c.code}, nil
}

func (c *mockClient) CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error {
	c.CountCopyToContainer++
	return nil
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
			[]types.Container{{Names: []string{"/name"}, State: "running"}},
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
			[]types.Container{{Names: []string{"/name"}}},
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
	tarFile = func(string, string) (io.ReadCloser, error) {
		return mockReader(1), nil
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

func TestRunContainer(t *testing.T) {
	tests := []struct {
		name,
		containerName,
		args string
		counter     clientCounter
		containers  []types.Container
		err         error
		out         string
		code        int
		reader      *bufio.Reader
		insertFiles []string
	}{
		{
			"couldn't find container",
			"name",
			"",
			clientCounter{
				CountContainerList: 1,
			},
			[]types.Container{},
			errors.New("couldn't find container name"),
			"",
			0,
			&bufio.Reader{},
			[]string{},
		},
		{
			"run successfully",
			"name",
			"arg",
			clientCounter{
				CountContainerList:        1,
				CountContainerExecAttach:  1,
				CountContainerExecCreate:  1,
				CountContainerExecInspect: 1,
				CountCopyToContainer:      3,
			},
			[]types.Container{
				{Names: []string{"/name"}},
			},
			nil,
			"out",
			0,
			bufio.NewReader(bytes.NewBuffer([]byte{2, 0, 0, 0, 0, 0, 0, 3, 'o', 'u', 't'})),
			[]string{"ff"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tarFile = func(string, string) (io.ReadCloser, error) {
				return mockReader(1), nil
			}
			client := &mockClient{containers: test.containers, code: test.code, reader: test.reader}
			dockerClient = client
			out, code, err := RunContainer(test.containerName, test.args, &config.Target{}, test.insertFiles)
			if fmt.Sprintf("%v", test.err) != fmt.Sprintf("%v", err) {
				t.Errorf("wanted error(%v), got(%v)", test.err, err)
			}
			if test.out != out {
				t.Errorf("wanted out(%v), got(%v)", test.out, out)
			}
			if test.code != code {
				t.Errorf("wanted code(%d), got(%d)", test.code, code)
			}
			if diff := cmp.Diff(test.counter, client.clientCounter); diff != "" {
				t.Errorf("method call counter mismatch (-want +got): %s", diff)
			}
		})
	}
}
