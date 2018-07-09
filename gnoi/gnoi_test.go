package gnoi

import (
	"testing"

	"google.golang.org/grpc/reflection"
)

func TestServer(t *testing.T) {
	s, err := NewServer(nil)
	if err != nil {
		t.Fatal("failed to Create Server:", err)
	}
	g, err := s.PrepareEncrypted()
	if err != nil {
		t.Fatal("failed to prepare encrypted gRPC Server:", err)
	}

	// g = s.PrepareAuthenticated()
	s.RegCertificateManagement(g)
	reflection.Register(g)
}
