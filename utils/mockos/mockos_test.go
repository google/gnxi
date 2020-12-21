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
					Hash:                  []byte{127, 63, 248, 60, 103, 96, 72, 104, 168, 195, 34, 35, 147, 150, 74, 165, 142, 151, 158, 106, 191, 249, 103, 224, 157, 41, 222, 176, 116, 158, 103, 96},
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
			want: []byte{127, 63, 248, 60, 103, 96, 72, 104, 168, 195, 34, 35, 147, 150, 74, 165, 142, 151, 158, 106, 191, 249, 103, 224, 157, 41, 222, 176, 116, 158, 103, 96},
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
			want: []byte{54, 11, 186, 215, 225, 58, 244, 159, 230, 156, 31, 18, 99, 184, 130, 41, 91, 84, 157, 36, 113, 106, 199, 251, 64, 167, 214, 245, 87, 210, 225, 165},
		},
	}
	for _, test := range tests {
		hash := calcHash(test.os)
		if !bytes.Equal(test.want, hash) {
			t.Errorf("Want %v, got %v", test.want, hash)
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
					Hash:                  []byte{0xd3, 0x73, 0x78, 0x56, 0xdd, 0xc4, 0x2e, 0xde, 0x11, 0x64, 0xf9, 0x41, 0x7d, 0xf6, 0x2d, 0xfc, 0x8c, 0x1a, 0x7f, 0x41, 0x7c, 0x2d, 0x08, 0xf5, 0x74, 0x4b, 0xc8, 0xee, 0xd6, 0x51, 0xf8, 0xfa},
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
