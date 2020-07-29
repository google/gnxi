package config

import (
	"errors"

	"github.com/spf13/viper"
)

// Device stores connection details of a target.
type Device struct {
	Address string `mapstructure:"address"`
	Ca      string `mapstructure:"ca"`
	CaKey   string `mapstructure:"cakey"`
}

// SetTarget adds any new target to the target history.
func SetTarget(targetName, targetAddress, ca, caKey string) error {
	err := prepareTarget(targetName, targetAddress, ca, caKey)
	if err != nil {
		return err
	}
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

func prepareTarget(targetName, targetAddress, ca, caKey string) error {
	devices := GetDevices()
	if devices == nil {
		devices = map[string]Device{}
	}
	if targetName == "" {
		if len(devices) > 0 {
			return nil
		}
		return errors.New("No targets in history and no target specified")
	}
	if _, exists := devices[targetName]; !exists {
		if targetAddress == "" || ca == "" || caKey == "" {
			return errors.New("Device not found")
		}
		devices[targetName] = Device{
			Address: targetAddress,
			Ca:      ca,
			CaKey:   caKey,
		}
	} else {
		device := devices[targetName]
		if targetAddress != "" {
			device.Address = targetAddress
		}
		if ca != "" {
			device.Ca = ca
		}
		if caKey != "" {
			device.CaKey = caKey
		}
		devices[targetName] = device
	}
	viper.Set("targets.last_target", targetName)
	viper.Set("targets.devices", devices)
	return nil
}
