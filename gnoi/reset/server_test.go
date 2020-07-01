package reset

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/gnxi/gnoi/reset/pb"
)

type test struct {
	name     string
	request  *pb.StartRequest
	settings Settings
}

func makeTests() []test {
	reqs := []pb.StartRequest{
		{ZeroFill: true},
		{ZeroFill: true, FactoryOs: true},
		{FactoryOs: true},
		{},
	}
	sets := []Settings{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}
	t := []test{}
	for _, set := range sets {
		for _, req := range reqs {
			name := fmt.Sprintf(
				"ZF:{Given:%v, Unsupported:%v},OS:{Given:%v,Unsupported:%v}",
				req.ZeroFill,
				set.ZeroFillUnsupported,
				req.FactoryOs,
				set.FactoryOSUnsupported,
			)
			t = append(t, test{name: name, request: &req, settings: set})
		}
	}
	return t
}

func TestStart(t *testing.T) {
	tests := makeTests()
	for _, test := range tests {
		s := NewServer(&test.settings)
		t.Run(test.name, func(t *testing.T) {
			resp, err := s.Start(context.Background(), test.request)
			if err != nil {
				t.Fatalf("ops")
			}
			switch response := resp.Response.(type) {
			case *pb.StartResponse_ResetSuccess:
				if test.settings.ZeroFillUnsupported && test.request.ZeroFill || test.settings.FactoryOSUnsupported && test.request.FactoryOs {
					t.Errorf(
						"Error occured on case: \nSettings{errIfZero:%v,osUnsupported:%v}\nRequest{ZeroFill:%v,FactoryOs:%v} \nResponseSuccess{%v},",
						test.settings.ZeroFillUnsupported,
						test.settings.FactoryOSUnsupported,
						test.request.ZeroFill,
						test.request.FactoryOs,
						response)
				}
			case *pb.StartResponse_ResetError:
				if response.ResetError.ZeroFillUnsupported != (test.settings.ZeroFillUnsupported && test.request.ZeroFill) ||
					response.ResetError.FactoryOsUnsupported != (test.settings.FactoryOSUnsupported && test.request.FactoryOs) {
					t.Errorf(
						"Error occured on case: \nSettings{errIfZero:%v,osUnsupported:%v}\nRequest{ZeroFill:%v,FactoryOs:%v} \nResponseError{%v},",
						test.settings.ZeroFillUnsupported,
						test.settings.FactoryOSUnsupported,
						test.request.ZeroFill,
						test.request.FactoryOs,
						response)
				}
			}
		})
	}
}
