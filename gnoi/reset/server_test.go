package reset

import (
	"context"
	"testing"

	"github.com/google/gnxi/gnoi/reset/pb"
)

func TestStart(t *testing.T) {
	testSettings := []Settings{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}

	testRequests := []pb.StartRequest{
		{ZeroFill: true},
		{ZeroFill: true, FactoryOs: true},
		{FactoryOs: true},
		{},
	}

	for _, test := range testSettings {
		s := NewServer(&test)
		for _, testReq := range testRequests {
			resp, _ := s.Start(context.Background(), &testReq)
			switch response := resp.Response.(type) {
			case *pb.StartResponse_ResetSuccess:
				if test.errorIfZero && testReq.ZeroFill || test.osUnsupported && testReq.FactoryOs {
					t.Errorf(
						"Error occured on case: \nSettings{errIfZero:%v,osUnsupported:%v}\nRequest{ZeroFill:%v,FactoryOs:%v} \nResponseSuccess{%v},",
						test.errorIfZero,
						test.osUnsupported,
						testReq.ZeroFill,
						testReq.FactoryOs,
						response)
				}
			case *pb.StartResponse_ResetError:
				if response.ResetError.ZeroFillUnsupported != (test.errorIfZero && testReq.ZeroFill) ||
					response.ResetError.FactoryOsUnsupported != (test.osUnsupported && testReq.FactoryOs) {
					t.Errorf(
						"Error occured on case: \nSettings{errIfZero:%v,osUnsupported:%v}\nRequest{ZeroFill:%v,FactoryOs:%v} \nResponseError{%v},",
						test.errorIfZero,
						test.osUnsupported,
						testReq.ZeroFill,
						testReq.FactoryOs,
						response)
				}
			}
		}
	}
}
