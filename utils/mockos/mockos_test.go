package mockos

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/utils/mockos/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
		os *OS
	}{
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "1.0a",
					Cookie:                "yyhnueaolrkoehvxcphhkxgvcilzppoh",
					Padding:               []byte("These are some bits to pad things out"),
					Incompatible:          false,
					ActivationFailMessage: "",
					Hash:                  []byte{236, 238, 38, 29, 175, 162, 142, 132, 150, 19, 79, 233, 230, 165, 179, 35},
				},
			},
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
		},
		{
			os: nil,
		},
	}
	for _, test := range tests {
		var serializedOS []byte
		if test.os != nil {
			serializedOS, _ = proto.Marshal(&test.os.MockOS)
		} else {
			serializedOS = make([]byte, 1000000)
			rand.Read(serializedOS)
		}
		bb := new(bytes.Buffer)
		bb.Write(serializedOS)
		os := ValidateOS(bb)
		if diff := cmp.Diff(test.os, os, cmpopts.IgnoreFields(pb.MockOS{}, "XXX_sizecache")); diff != "" {
			t.Errorf("ValidateOS(): (-want +got):\n%s", diff)
		}
	}
}
