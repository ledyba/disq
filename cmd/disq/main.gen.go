package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func buildAt() string {
	return "2017/12/19 22:26:41"
}

func gitRev() string {
	data, err := base64.StdEncoding.DecodeString("Y29tbWl0IDk2MzIwNTQwYzAzODdmNzc5ZmJjZmUyMzViODk4MWFjYmRmNTUyNzgKQXV0aG9yOiBwc2kgPHBzaUBsZWR5YmEub3JnPgpEYXRlOiAgIFR1ZSBEZWMgMTkgMDg6MTI6MDkgMjAxNyArMDkwMAoKICAgIHJlYWRtZQo=")
	if err != nil {
		return fmt.Sprintf("<an error occured while reading git rev: %v>", err)
	}
	if len(data) == 0 {
		return "<not available>"
	}
	return strings.TrimSpace(string(data))
}

