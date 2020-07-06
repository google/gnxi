package os

import (
	"context"
	"testing"

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

type activateRPC func(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error)

type mockClient struct {
	pb.OSClient
	activate activateRPC
}

type activateTest struct {
	name   string
	client *mockClient
	want   *ActivateError
}

func (c *mockClient) Activate(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error) {
	return c.activate(ctx, in, opts...)
}

func activateErrorRPC(errType pb.ActivateError_Type, detail string) activateRPC {
	return func(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error) {
		return &pb.ActivateResponse{
			Response: &pb.ActivateResponse_ActivateError{
				ActivateError: &pb.ActivateError{Type: errType, Detail: detail},
			}}, nil
	}
}

func activateSuccessRPC(ctx context.Context, in *pb.ActivateRequest, opts ...grpc.CallOption) (*pb.ActivateResponse, error) {
	return &pb.ActivateResponse{
		Response: &pb.ActivateResponse_ActivateOk{}}, nil
}

func generateActivateTests() []activateTest {
	tests := []activateTest{
		{
			"Success",
			&mockClient{activate: activateSuccessRPC},
			nil,
		},
		{
			"Unspecified",
			&mockClient{activate: activateErrorRPC(pb.ActivateError_UNSPECIFIED, "detail")},
			&ActivateError{ErrType: ActivateUnspecified, Detail: "detail"},
		},
		{
			"Non Existent Version",
			&mockClient{activate: activateErrorRPC(pb.ActivateError_NON_EXISTENT_VERSION, "detail")},
			&ActivateError{ErrType: ActivateNonExistentVersion, Detail: "detail"},
		},
	}
	return tests
}

func TestActivate(t *testing.T) {
	activateTests := generateActivateTests()
	for _, test := range activateTests {
		t.Run(test.name, func(t *testing.T) {
			client := Client{client: test.client}
			got := client.Activate(context.Background(), "version")
			if got == nil && test.want == nil {
				return
			}
			if err := got.(*ActivateError); err.ErrType != test.want.ErrType || err.Detail != test.want.Detail {
				t.Errorf("Wanted: ActivateError(%v), Got: ActivateError(%v)", *test.want, *err)
			}
		})
	}
}
