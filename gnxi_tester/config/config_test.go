package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestGetTests(t *testing.T) {
	wants := []map[string][]Test{
		{},
		{"gnoi_os": []Test{{
			Name:   "Compatible OS with Good Hash Install",
			Args:   map[string]string{"op": "install", "version": "&<version>", "os": "&<os_path>"},
			Wants:  `^$`,
			Prompt: []string{"version", "os_path"},
		}}},
		{
			"gnoi_cert": []Test{
				{
					Name:  "Rotate Certificate",
					Args:  map[string]string{"op": "rotate", "cert_id": "&<cert_id>"},
					Wants: "Rotate success",
				},
				{
					Name:     "Rotate Non-Existent Certificate",
					Args:     map[string]string{"op": "rotate", "cert_id": "&<non_existent_cert_id>"},
					MustFail: true,
					Wants:    "Failed Rotate",
					Prompt:   []string{"non_existent_cert_id"},
				},
			},
			"gnoi_reset": []Test{{
				Name:  "Resetting a Target Successfully",
				Args:  map[string]string{},
				Wants: `^$`,
			}},
		},
	}
	for _, want := range wants {
		viper.Set("tests", want)
		got := GetTests()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("GetTests(): (-got +want):\n%s", diff)
		}
	}
}
