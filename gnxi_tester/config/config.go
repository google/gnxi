package config

import (
	"path"

	log "github.com/golang/glog"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Init will read and if needed, initialize the config file.
func Init(filePath string) {
	if filePath != "" {
		home, err := homedir.Dir()
		if err != nil {
			log.Exitf("couldn't get home directory: %v", err)
		}
		filePath = path.Join(home, ".gnxi.yml")
	}
	setDefaults()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filePath)
	viper.SafeWriteConfig()
	viper.ReadInConfig()
}

// GetTests will return tests from viper store.
func GetTests() map[string][]Test {
	return viper.Get("tests").(map[string][]Test)
}
