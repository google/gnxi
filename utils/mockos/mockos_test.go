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

func TestCheckHash(t *testing.T) {
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

func TestCalcHash(t *testing.T) {
	tests := []struct {
		os   *OS
		want []byte
	}{
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "1.0a",
					Cookie:                "yyhnueaolrkoehvxcphhkxgvcilzppoh",
					Padding:               []byte("These are some bits to pad things out"),
					Incompatible:          true,
					ActivationFailMessage: "This is a test activationFailMessage",
				},
			},
			want: []byte{252, 85, 174, 147, 254, 203, 37, 51, 0, 249, 102, 183, 117, 56, 23, 63},
		},
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "2.2b",
					Cookie:                "yyhnueaolrasdfasdgdfsgdflkjsakfsdkoh",
					Padding:               []byte("These are some bits to pad things out a lot!"),
					Incompatible:          true,
					ActivationFailMessage: "This is a test activationFailMessage",
				},
			},
			want: []byte{98, 49, 38, 77, 248, 23, 48, 91, 68, 99, 25, 226, 238, 224, 47, 213},
		},
	}
	for _, test := range tests {
		hash := calcHash(test.os)
		if !bytes.Equal(test.want, hash) {
			t.Errorf("Want %x, got %x", test.want, hash)
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

func TestPackageOS(t *testing.T) {
	tests := []struct {
		os *OS
	}{
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:               "1.0a",
					Cookie:                "cookiestring",
					Incompatible:          true,
					ActivationFailMessage: "This is a test activationFailMessage",
					Hash:                  []byte{122, 63, 198, 213, 205, 32, 172, 235, 80, 248, 51, 116, 45, 185, 200, 210},
				},
			},
		},
	}
	for _, test := range tests {
		got := &OS{MockOS: pb.MockOS{}}
		serializedOS, _ := packageOS(test.os.MockOS.Version, "0", test.os.MockOS.ActivationFailMessage, test.os.MockOS.Incompatible)
		_ = proto.Unmarshal(serializedOS, &got.MockOS)
		if diff := cmp.Diff(test.os, got); diff != "" {
			t.Errorf("ValidateOS(): (-want +got):\n%s", diff)
		}
	}
}
