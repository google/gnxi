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
					Version:     "1.0a",
					Cookie:      "yyhnueaolrkoehvxcphhkxgvcilzppoh",
					Padding:     []byte("These are some bits to pad things out"),
					Unsupported: true,
					Hash:        []byte{104, 17, 158, 51, 77, 54, 186, 113, 151, 56, 105, 23, 200, 168, 255, 120},
				},
			},
			want: true,
		},
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:     "2.2b",
					Cookie:      "yyhnueaolrasdfasdgdfsgdflkjsakfsdkoh",
					Padding:     []byte("These are some bits to pad things out a lot!"),
					Unsupported: false,
					Hash:        []byte("Bad Hash"),
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
