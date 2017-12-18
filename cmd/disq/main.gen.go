package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func buildAt() string {
	return "2017/12/18 22:14:34"
}

func gitRev() string {
	data, err := base64.StdEncoding.DecodeString("Y29tbWl0IGM2MjlkMDBmOTQ2YTAwODczZmM3ZjEzY2QyZjBkMDU4YjNiZTY2ZjgKQXV0aG9yOiBwc2kgPHBzaUBsZWR5YmEub3JnPgpEYXRlOiAgIE1vbiBEZWMgMTEgMTM6MTM6MjQgMjAxNyArMDkwMAoKICAgIGluaXQK")
	if err != nil {
		return fmt.Sprintf("<an error occured while reading git rev: %v>", err)
	}
	if len(data) == 0 {
		return "<not available>"
	}
	return strings.TrimSpace(string(data))
}

