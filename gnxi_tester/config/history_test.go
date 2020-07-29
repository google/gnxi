package config

import (
	"errors"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/spf13/viper"
)

type Want struct {
	devices map[string]Device
	err     error
}

func TestPrepareTarget(t *testing.T) {
	tests := []struct {
		name          string
		targetName    string
		targetAddress string
		targetCA      string
		targetCAKey   string
		config        map[string]Device
		lastTarget    string
		want          Want
	}{
		{
			name: "No targets in history, no target specified",
			want: Want{
				err:     errors.New("No targets in history and no target specified"),
				devices: map[string]Device{},
			},
		},
		{
			name: "No target specified",
			config: map[string]Device{"myhost.com": Device{
				Address: "localhost:9339",
				Ca:      "ca.crt",
				CaKey:   "ca.key",
			}},
			lastTarget: "myhost.com",
			want: Want{
				err: nil,
				devices: map[string]Device{"myhost.com": Device{
					Address: "localhost:9339",
					Ca:      "ca.crt",
					CaKey:   "ca.key",
				}},
			},
		},
		{
			name: "Non-existent device",
			config: map[string]Device{"myhost.com": Device{
				Address: "localhost:9339",
				Ca:      "ca.crt",
				CaKey:   "ca.key",
			}},
			targetName: "nonexistenttarget",
			lastTarget: "myhost.com",
			want: Want{
				err: errors.New("Device not found"),
				devices: map[string]Device{"myhost.com": Device{
					Address: "localhost:9339",
					Ca:      "ca.crt",
					CaKey:   "ca.key",
				}},
			},
		},
		{
			name:          "Add new target",
			targetName:    "myhost.com",
			targetAddress: "localhost:9339",
			targetCA:      "ca.crt",
			targetCAKey:   "ca.key",
			want: Want{
				devices: map[string]Device{
					"myhost.com": {
						Address: "localhost:9339",
						Ca:      "ca.crt",
						CaKey:   "ca.key",
					},
				},
			},
		},
		{
			name:          "Update existing device's address",
			targetName:    "myhost.com",
			targetAddress: "newhost:9340",
			config: map[string]Device{"myhost.com": Device{
				Address: "localhost:9339",
				Ca:      "ca.crt",
				CaKey:   "ca.key",
			}},
			want: Want{
				devices: map[string]Device{
					"myhost.com": {
						Address: "newhost:9340",
						Ca:      "ca.crt",
						CaKey:   "ca.key",
					},
				},
			},
		},
		{
			name:       "Update existing device's ca",
			targetName: "myhost.com",
			targetCA:   "newca.crt",
			config: map[string]Device{"myhost.com": Device{
				Address: "localhost:9339",
				Ca:      "ca.crt",
				CaKey:   "ca.key",
			}},
			want: Want{
				devices: map[string]Device{
					"myhost.com": {
						Address: "localhost:9339",
						Ca:      "newca.crt",
						CaKey:   "ca.key",
					},
				},
			},
		},
		{
			name:        "Update existing device's ca_key",
			targetName:  "myhost.com",
			targetCAKey: "newca.key",
			config: map[string]Device{"myhost.com": Device{
				Address: "localhost:9339",
				Ca:      "ca.crt",
				CaKey:   "ca.key",
			}},
			want: Want{
				devices: map[string]Device{
					"myhost.com": {
						Address: "localhost:9339",
						Ca:      "ca.crt",
						CaKey:   "newca.key",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var devices map[string]Device
			viper.Set("targets.last_target", test.lastTarget)
			viper.Set("targets.devices", test.config)
			err := prepareTarget(test.targetName, test.targetAddress, test.targetCA, test.targetCAKey)
			viper.UnmarshalKey("targets.devices", &devices)
			got := Want{
				devices: devices,
				err:     err,
			}
			if diff := pretty.Compare(test.want, got); diff != "" {
				t.Errorf("prepareTarget(%s, %s, %s, %s): (-want +got)\n%s", test.targetName, test.targetAddress, test.targetCA, test.targetCAKey, diff)
			}
		})
	}
}
