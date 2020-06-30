package reset

import (
	"context"
	"errors"
	"log"

	"github.com/google/gnxi/gnoi/reset/pb"
	"google.golang.org/grpc"
)

// Client handles requesting a Factory Reset.
type Client struct {
	client pb.FactoryResetClient
}

// NewClient initializes a FactoryReset Client.
func NewClient(c *grpc.ClientConn) *Client {
	return &Client{client: pb.NewFactoryResetClient(c)}
}

// ResetTarget requests a factory reset.
func (c *Client) ResetTarget(ctx context.Context, zeroFill, rollbackOS bool) error {
	out, err := c.client.Start(ctx, &pb.StartRequest{
		FactoryOs: rollbackOS,
		ZeroFill:  zeroFill,
	})
	if err != nil {
		log.Println("Error calling Start service:", err)
		return err
	}
	return CheckResponse(out)
}

// CheckResponse checks for errors.
func CheckResponse(res *pb.StartResponse) error {
	log.Println(res)
	switch res.Response.(type) {
	case *pb.StartResponse_ResetSuccess:
		return nil
	case *pb.StartResponse_ResetError:
		resErr := res.GetResetError()
		if resErr.FactoryOsUnsupported {
			log.Println("Factory OS Rollback Unsupported")
			return errors.New("Factory OS Rollback Unsupported")
		}
		if resErr.ZeroFillUnsupported {
			log.Println("Zero Filling Persistent Storage Unsupported")
			return errors.New("Zero Fill Unsupported")
		}
		if resErr.Other {
			log.Println(resErr.Detail)
			return errors.New(resErr.Detail)
		}
	}
	return nil
}
