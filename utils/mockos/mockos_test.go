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
					Version:   "1.0a",
					Cookie:    "yyhnueaolrkoehvxcphhkxgvcilzppoh",
					Padding:   []byte("These are some bits to pad things out"),
					Supported: true,
					Hash:      []byte{104, 17, 158, 51, 77, 54, 186, 113, 151, 56, 105, 23, 200, 168, 255, 120},
				},
			},
			want: true,
		},
		{
			os: &OS{
				MockOS: pb.MockOS{
					Version:   "2.2b",
					Cookie:    "yyhnueaolrasdfasdgdfsgdflkjsakfsdkoh",
					Padding:   []byte("These are some bits to pad things out a lot!"),
					Supported: false,
					Hash:      []byte("Bad Hash"),
				},
			},
			want: false,
		},
	}

	for _, test := range tests {
		got := test.os.CheckHash()
		if got != test.want {
			t.Errorf("Expected %v, got %v", test.want, got)
		}
	}
}
