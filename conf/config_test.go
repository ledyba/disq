package conf

import (
	"testing"

	util "github.com/ledyba/disq/util-test"
)

func TestEmptyConfig(t *testing.T) {
	dat := util.ReadAll(t, "../config-sample.json")
	actual, err := Load(dat)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(actual.DNS.Networks) == 0 {
		t.Errorf("We got empty DNS allowed network list.")
	}
	if len(actual.V4Networks) == 0 {
		t.Errorf("We got empty v4 network list.")
	}
	if len(actual.Machines) == 0 {
		t.Errorf("We got empty machine list.")
	}
}
