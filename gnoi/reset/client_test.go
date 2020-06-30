package reset

import (
	"testing"

	"github.com/google/gnxi/gnoi/reset/pb"
)

func TestStart(t *testing.T) {
	t.Run("ResetSuccess", func(t *testing.T) {
		res := initializeResponse(false, false, false, "")
		err := CheckResponse(res)
		if err != nil {
			t.Errorf("Expected no errors, got %v", err)
		}
	})

	t.Run("ResetError OS Rollback Impossible", func(t *testing.T) {
		res := initializeResponse(true, false, false, "")
		err := CheckResponse(res)
		if err == nil {
			t.Error("Expected OS unsupported error, no error returned")
		}
	})
	t.Run("ResetError Zero Fill Impossible", func(t *testing.T) {
		res := initializeResponse(false, true, false, "")
		err := CheckResponse(res)
		if err == nil {
			t.Error("Expected Zero Fill unsupported error, no error returned")
		}
	})
	t.Run("Unspecified Error Response", func(t *testing.T) {
		res := initializeResponse(false, false, true, "Unspecified Test Error")
		err := CheckResponse(res)
		if err == nil {
			t.Error("Expected Unspecified error, no error returned")
		}
	})
}

func initializeResponse(factoryOSUnsupported, zeroFillUnsupported, other bool, details string) *pb.StartResponse {
	res := &pb.StartResponse{}
	if factoryOSUnsupported || zeroFillUnsupported || other {
		resetError := &pb.ResetError{
			FactoryOsUnsupported: factoryOSUnsupported,
			ZeroFillUnsupported:  zeroFillUnsupported,
			Other:                other,
			Detail:               details,
		}
		res.Response = &pb.StartResponse_ResetError{resetError}
	} else {
		resetSuccess := &pb.ResetSuccess{}
		res.Response = &pb.StartResponse_ResetSuccess{resetSuccess}
	}
	return res
}
