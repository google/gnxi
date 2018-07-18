/* Copyright 2018 Google Inc.

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

package gnoi

import (
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	conString := "127.0.0.1:4456"

	s, err := NewServer(nil, nil)
	if err != nil {
		t.Fatal("failed to Create Server:", err)
	}

	g := s.PrepareEncrypted()
	s.RegCertificateManagement(g)
	listen, err := net.Listen("tcp", conString)
	if err != nil {
		t.Fatal("server failed to listen:", err)
	}
	go g.Serve(listen)
	g.GracefulStop()
	listen.Close()

	g = s.PrepareAuthenticated()
	s.RegCertificateManagement(g)
	listen, err = net.Listen("tcp", conString)
	if err != nil {
		t.Fatal("server failed to listen:", err)
	}
	go g.Serve(listen)
	g.GracefulStop()
	listen.Close()
}
