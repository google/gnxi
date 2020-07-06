package os

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/gnxi/gnoi/os/pb"
)

var (
	server = initializeServer()
)

func initializeServer() *Server {
	srv := NewServer("1")
	srv.manager.Install("1.0.0a")
	return srv
}
func TestActivate(t *testing.T) {
	tests := []struct {
		request *pb.ActivateRequest
		want    *pb.ActivateResponse
	}{
		{
			request: &pb.ActivateRequest{
				Version: "1.0.0a",
			},
			want: &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateOk{}},
		},
		{
			request: &pb.ActivateRequest{
				Version: "99.0a",
			},
			want: &pb.ActivateResponse{Response: &pb.ActivateResponse_ActivateError{
				ActivateError: &pb.ActivateError{Type: 1},
			}},
		},
	}
	for _, test := range tests {
		got, _ := server.Activate(context.Background(), test.request)
		diff := reflect.DeepEqual(test.want, got)
		if !diff {
			t.Errorf("Activate(context.Background(), %s): Expected %v, Got %v", test.request, test.want, got)
		}
	}
}
