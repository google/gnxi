package config

import (
	"errors"

	"github.com/spf13/viper"
)

const defaultPort = ":9339"

// SetTarget adds any new target to the target history.
func SetTarget(targetName, targetAddress string) error {
	devices := GetDevices()
	if targetName == "" {
		if len(devices) > 0 {
			return nil
		}
		return errors.New("No targets in history and no target specified")
	}
	if devices[targetName] == "" {
		if targetAddress == "" {
			devices[targetName] = defaultPort
		} else {
			devices[targetName] = targetAddress
		}
	}
	viper.Set("targets.last_target", targetName)
	viper.Set("targets.devices", devices)
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}
