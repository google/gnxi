package config

import (
	"time"

	"github.com/spf13/viper"
)

// Test represents a single set of inputs and expected outputs.
type Test struct {
	Args     string        `mapstructure:"args"`
	MustFail bool          `mapstructure:"must_fail"`
	Wait     time.Duration `mapstructure:"wait"`
	Wants    string        `mapstructure:"wants"`
}

func setDefaults() {
	viper.SetDefault("tests", map[string][]Test{})
}
