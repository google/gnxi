package os

import (
	"context"
	"fmt"

	"github.com/google/gnxi/gnoi/os/pb"
	"google.golang.org/grpc"
)

// Client handles requesting OS RPCs.
type Client struct {
	client pb.OSClient
}

// ActivateErrorType represents the Type enum in ActivateError.
type ActivateErrorType string

// Enum representing possible ActivateErrorTypes.
const (
	ActivateUnspecified        ActivateErrorType = "UNSPECIFIED"
	ActivateNonExistentVersion                   = "NON_EXISTENT_VERSION"
)

// ActivateError represents an error returned by the Activate RPC.
type ActivateError struct {
	error
	ErrType ActivateErrorType
	Detail  string
}

func (e *ActivateError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrType, e.Detail)
}

// NewClient returns a new OS service client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewOSClient(c)}
}

// Activate invokes the Activate RPC for the OS service.
func (c *Client) Activate(ctx context.Context, version string) error {
	out, err := c.client.Activate(ctx, &pb.ActivateRequest{Version: version})
	if err != nil {
		return err
	}
	switch out.Response.(type) {
	case *pb.ActivateResponse_ActivateOk:
		return nil
	case *pb.ActivateResponse_ActivateError:
		res := out.GetActivateError()
		errType := ActivateErrorType(res.GetType().String())
		switch errType {
		case ActivateUnspecified:
			fallthrough
		case ActivateNonExistentVersion:
			return &ActivateError{
				ErrType: errType,
				Detail:  res.GetDetail(),
			}
		default:
			return fmt.Errorf("Unknown ActivateError type: %s", errType)
		}
	}
	return nil
}
