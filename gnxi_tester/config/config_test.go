package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func TestGetDevices(t *testing.T) {
	tests := []map[string]string{
		{},
		{"mydevice.com": "localhost:9339"},
		{
			"mydevice.com":      "localhost:9339",
			"anotherdevice.com": ":9400",
		},
	}
	for _, cfg := range tests {
		viper.Set("targets.devices", cfg)
		got := GetDevices()
		if diff := cmp.Diff(cfg, got); diff != "" {
			t.Errorf("GetDevices(): (-want +got):\n%s", diff)
		}
	}
}
