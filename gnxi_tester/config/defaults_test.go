package config

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/spf13/viper"
)

func TestGenerateTestCases(t *testing.T) {
	got := generateTestCases()
	if got == nil {
		t.Errorf("Tests Not Defined!")
	}
	if got["gnoi_os"] == nil {
		t.Errorf("OS Tests Not Defined!")
	}
	if got["gnoi_cert"] == nil {
		t.Errorf("Certificate Tests Not Defined!")
	}
	if got["gnoi_reset"] == nil {
		t.Errorf("Reset Tests Not Defined!")
	}
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	var got map[string][]Test
	want := generateTestCases()
	setDefaults()
	if err := viper.UnmarshalKey("tests", &got); err != nil {
		t.Errorf("Error getting tests: %v", err)
	}
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("GetTests(): (-want +got):\n%s", diff)
	}
}
