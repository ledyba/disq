package util_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Tester interface {
	Fatal(args ...interface{})
}

func ReadAll(t Tester, relpath string) []byte {
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
