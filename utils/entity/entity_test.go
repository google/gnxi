package entity

import (
	"path/filepath"
	"testing"
)

func testPath(name string) string {
	return filepath.Join("testData", name)
}

func TestEntity(t *testing.T) {
	root, err := CreateSelfSigned("root")
	if err != nil {
		t.Fatal("CreateSelfSigned(root):", err)
	}

	clientCA, err := CreateSignedCA("clientCA", root)
	if err != nil {
		t.Fatal("CreateSignedCA(clientCA):", err)
	}

	targetCA, err := CreateSignedCA("targetCA", root)
	if err != nil {
		t.Fatal("CreateSignedCA(targetCA):", err)
	}

	client, err := CreateSigned("client", clientCA)
	if err != nil {
		t.Fatal("CreateSigned(client):", err)
	}

	target, err := CreateSigned("target", targetCA)
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
