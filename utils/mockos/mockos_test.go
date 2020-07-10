package mockos

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
	osPb "github.com/google/gnxi/gnoi/os/pb"
	"github.com/google/gnxi/utils/mockos/pb"
	"github.com/kylelemons/godebug/pretty"
)

func TestHashValidation(t *testing.T) {
	tests := []struct {
		os   *OS
		want bool
	}{
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "1.0a",
					Cookie:                "yyhnueaolrkoehvxcphhkxgvcilzppoh",
					Padding:               []byte("These are some bits to pad things out"),
					Incompatible:          true,
					ActivationFailMessage: "This is a test activationFailMessage",
					Hash:                  []byte{252, 85, 174, 147, 254, 203, 37, 51, 0, 249, 102, 183, 117, 56, 23, 63},
				},
			},
			want: true,
		},
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "2.2b",
					Cookie:                "yyhnueaolrasdfasdgdfsgdflkjsakfsdkoh",
					Padding:               []byte("These are some bits to pad things out a lot!"),
					Incompatible:          true,
					ActivationFailMessage: "This is a test activationFailMessage",
					Hash:                  []byte("Bad Hash"),
				},
			},
		},
	}

	for _, test := range tests {
		got := test.os.CheckHash()
		if got != test.want {
			t.Errorf("Want %v, Got %v", test.want, got)
		}
	}
}

func TestValidateOS(t *testing.T) {
	tests := []struct {
		os   *OS
		want *osPb.InstallResponse_InstallError
	}{
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "1.0a",
					Cookie:                "yyhnueaolrkoehvxcphhkxgvcilzppoh",
					Padding:               []byte("These are some bits to pad things out"),
					Incompatible:          false,
					ActivationFailMessage: "This is a test activationFailMessage",
					Hash:                  []byte{236, 238, 38, 29, 175, 162, 142, 132, 150, 19, 79, 233, 230, 165, 179, 35},
				},
			},
			want: nil,
		},
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "2.2b",
					Cookie:                "yyhnueaolrasdfasdgdfsgdflkjsakfsdkoh",
					Padding:               []byte("These are some bits to pad things out a lot!"),
					Incompatible:          false,
					ActivationFailMessage: "This is a test activationFailMessage",
					Hash:                  []byte("Bad Hash"),
				},
			},
			want: &osPb.InstallResponse_InstallError{&osPb.InstallError{Type: osPb.InstallError_INTEGRITY_FAIL}},
		},
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "1.0a",
					Cookie:                "yyhnueaolrkoehvxcphhkxgvcilzppoh",
					Padding:               []byte("These are some bits to pad things out"),
					Incompatible:          true,
					ActivationFailMessage: "This is a test activationFailMessage",
					Hash:                  []byte{252, 85, 174, 147, 254, 203, 37, 51, 0, 249, 102, 183, 117, 56, 23, 63},
				},
			},
			want: &osPb.InstallResponse_InstallError{&osPb.InstallError{Type: osPb.InstallError_INCOMPATIBLE, Detail: "Unsupported OS Version"}},
		},
	}
	for _, test := range tests {
		serializedOS, _ := proto.Marshal(&test.os.MockOS)
		bb := new(bytes.Buffer)
		bb.Write(serializedOS)
		_, _, errRes := ValidateOS(bb)
		if diff := pretty.Compare(test.want, errRes); diff != "" {
			t.Errorf("ValidateOS(): (-want +got):\n%s", diff)
		}
	}
}
