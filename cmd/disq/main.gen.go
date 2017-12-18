package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func buildAt() string {
	return "2017/12/18 22:57:20"
}

func gitRev() string {
	data, err := base64.StdEncoding.DecodeString("Y29tbWl0IDBlMDU5OTM5OWQ0OWI4MzFkMTIwM2ZhOGQxZDA5ZTkxYWVhMWE0MGIKQXV0aG9yOiBwc2kgPHBzaUBsZWR5YmEub3JnPgpEYXRlOiAgIFR1ZSBEZWMgMTkgMDc6MTQ6NTEgMjAxNyArMDkwMAoKICAgIERhZW1vbuOBqOOBl+OBpuOBruOBneOCjOOBo+OBveOBleOCkueNsuW+l+OBl+OBpuOBjeOBnwo=")
	if err != nil {
		return fmt.Sprintf("<an error occured while reading git rev: %v>", err)
	}
	if len(data) == 0 {
		return "<not available>"
	}
	return strings.TrimSpace(string(data))
}

