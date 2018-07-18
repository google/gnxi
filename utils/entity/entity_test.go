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

package entity

import (
	"crypto/x509"
	"path/filepath"
	"testing"
)

func testPath(name string) string {
	return filepath.Join("testData", name)
}

func TestEntity(t *testing.T) {
	root, err := CreateSelfSigned("root", nil)
	if err != nil {
		t.Fatal("CreateSelfSigned(root):", err)
	}

	clientCA, err := CreateSignedCA("clientCA", nil, root)
	if err != nil {
		t.Fatal("CreateSignedCA(clientCA):", err)
	}

	targetCA, err := CreateSignedCA("targetCA", nil, root)
	if err != nil {
		t.Fatal("CreateSignedCA(targetCA):", err)
	}

	client, err := CreateSigned("client", nil, clientCA)
	if err != nil {
		t.Fatal("CreateSigned(client):", err)
	}

	target, err := CreateSigned("target", nil, targetCA)
	if err != nil {
		t.Fatal("CreateSigned(client):", err)
	}

	tests := []struct {
		fail       bool
		child      *Entity
		parent     *Entity
		childName  string
		parentName string
	}{
		{false, client, clientCA, "client", "clientCA"},
		{false, target, targetCA, "target", "targetCA"},
		{false, targetCA, root, "targetCA", "root"},
		{false, clientCA, root, "clientCA", "root"},
		{true, client, root, "client", "root"},
		{true, target, root, "target", "root"},
		{true, target, clientCA, "target", "clientCA"},
		{true, client, targetCA, "client", "targetCA"},
		{true, client, target, "client", "target"},
		{true, target, client, "target", "client"},
	}
	for _, test := range tests {
		if err := test.child.SignedBy(test.parent); (err != nil) != test.fail {
			t.Errorf("%s not signed by %s", test.childName, test.parentName)
		}
	}
}

func TestEntityFromFile(t *testing.T) {
	root, err := FromFile(testPath("root.crt"), testPath("root.key"))
	if err != nil {
		t.Fatal("EntityFromFile(root):", err)
	}

	target, err := FromFile(testPath("target.crt"), testPath("target.key"))
	if err != nil {
		t.Fatal("EntityFromFile(target):", err)
	}

	client, err := FromFile(testPath("client.crt"), testPath("client.key"))
	if err != nil {
		t.Fatal("EntityFromFile(client):", err)
	}

	tests := []struct {
		fail       bool
		child      *Entity
		parent     *Entity
		childName  string
		parentName string
	}{
		{false, client, root, "client", "root"},
		{false, target, root, "target", "root"},
		{true, client, target, "client", "target"},
		{true, target, client, "target", "client"},
	}
	for _, test := range tests {
		if err := test.child.SignedBy(test.parent); (err != nil) != test.fail {
			t.Errorf("%s not signed by %s", test.childName, test.parentName)
		}
	}
}

func TestFromSigningRequest(t *testing.T) {
	e, err := NewEntity(Template("requester"), nil)
	if err != nil {
		t.Fatal("failed to create an Entity:", err)
	}
	csrDER, err := e.SigningRequest()
	if err != nil {
		t.Fatal("failed to create a CSR:", err)
	}

	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		t.Fatal("failed to parse a CSR in DER enconding:", err)
	}

	if err = csr.CheckSignature(); err != nil {
		t.Fatal("CSR signature check failed:", err)
	}

	ne, err := FromSigningRequest(csr)
	if err != nil {
		t.Fatal("failed to create an Entity from a CSR:", err)
	}

	root, err := CreateSelfSigned("root", nil)
	if err != nil {
		t.Fatal("failed to create a Self Signed Certificate:", err)
	}

	if err := ne.SignWith(root); err != nil {
		t.Fatal("failed to sign a CSR generated Entity:", err)
	}
}
