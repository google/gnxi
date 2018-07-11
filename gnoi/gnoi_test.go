package gnoi

import (
	"net"
	"testing"
	"time"

	"google.golang.org/grpc/reflection"
)

func TestServer(t *testing.T) {
	s, err := NewServer(nil, nil)
	if err != nil {
		t.Fatal("failed to Create Server:", err)
	}
	g := s.PrepareEncrypted()

	// g = s.PrepareAuthenticated()
	s.RegCertificateManagement(g)
	reflection.Register(g)

	conString := "127.0.0.1:4456"
	listen, err := net.Listen("tcp", conString)
	if err != nil {
		t.Fatal("server failed to listen:", err)
	}
	go g.Serve(listen)
	defer g.GracefulStop()
	time.Sleep(time.Second)

}
