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

package cert

import (
	"context"

	"github.com/google/gnxi/gnoi/cert/pb"
	"google.golang.org/grpc"
)

type rotateRequestMap struct {
	req  *pb.RotateCertificateRequest
	resp *pb.RotateCertificateResponse
}
type rotateClient struct {
	pb.CertificateManagement_RotateClient
	reqMap  []*rotateRequestMap
	i       int
	recv    chan int
	recvErr chan *pb.RotateCertificateResponse
}

type mockClient struct {
	pb.CertificateManagementClient
	rotate *rotateClient
}

func (c *rotateClient) Send(req *pb.RotateCertificateRequest) error {
	return nil
}

func (c *rotateClient) Recv() (*pb.RotateCertificateResponse, error) {
	return nil, nil
}

func (c *mockClient) Rotate(ctx context.Context, opts ...grpc.CallOption) (pb.CertificateManagement_RotateClient, error) {
	return c.rotate, nil
}
