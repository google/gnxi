package config

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/spf13/viper"
)

func TestGenerateTestCases(t *testing.T) {
	got, order := generateTestCases()
	for _, testName := range order {
		if _, ok := got[testName]; !ok {
			t.Errorf("Wanted %s, but not retrieved!", testName)
		}
	}
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	var testsGot map[string][]Test
	var orderGot []string
	testsWanted, orderWanted := generateTestCases()
	want := struct {
		tests map[string][]Test
		order []string
	}{
		tests: testsWanted,
		order: orderWanted,
	}
	setDefaults()
	if err := viper.UnmarshalKey("tests", &testsGot); err != nil {
		t.Errorf("Error getting tests: %v", err)
	}
	if err := viper.UnmarshalKey("order", &orderGot); err != nil {
		t.Errorf("Error getting order: %v", err)
	}
	got := struct {
		tests map[string][]Test
		order []string
	}{
		tests: testsGot,
		order: orderGot,
	}
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("GetTests(): (-want +got):\n%s", diff)
	}
}
