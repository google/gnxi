package mockos

import (
	"testing"

	"github.com/google/gnxi/utils/mockos/pb"
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
