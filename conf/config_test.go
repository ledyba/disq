package conf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func readAll(t *testing.T, relpath string) []byte {
	cd, err := os.Getwd()
	if err != nil {
		t.Fatal("Error when calling getwd: ", err)
	}
	abspath := filepath.Join(cd, relpath)
	f, err := os.Open(abspath)
	if err != nil {
		t.Fatal("Error when opening: ", abspath, "err=", err)
	}
	defer f.Close()
	dat, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("Error when reading all ", abspath, "err=", err)
	}
	return dat
}

func TestEmptyConfig(t *testing.T) {
	dat := readAll(t, "../config-sample.json")
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
