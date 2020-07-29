package config

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestGetDefaults(t *testing.T) {
	got := generateTestCases()
	if got == nil {
		t.Errorf("Reset Tests Not Set!")
	}
	if got["gnoi_os"] == nil {
		t.Errorf("OS Tests Not Set!")
	}
	if got["gnoi_cert"] == nil {
		t.Errorf("Certificate Tests Not Set!")
	}
	if got["gnoi_reset"] == nil {
		t.Errorf("Reset Tests Not Set!")
	}
}

func TestSetDefaults(t *testing.T) {
	want := generateTestCases()
	setDefaults()
	got := GetTests()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("GetTests(): (-got +want):\n%s", diff)
	}
}
