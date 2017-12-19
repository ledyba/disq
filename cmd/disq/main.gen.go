package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func buildAt() string {
	return "2017/12/19 23:15:31"
}

func gitRev() string {
	data, err := base64.StdEncoding.DecodeString("Y29tbWl0IDhhMTExZGQ0MzVhMDYzMjM3YTRmYTFmM2UzZTI4Mzg5ZWNjMGFlZTMKQXV0aG9yOiBwc2kgPHBzaUBsZWR5YmEub3JnPgpEYXRlOiAgIFdlZCBEZWMgMjAgMDc6MjY6NTIgMjAxNyArMDkwMAoKICAgIGdyYWNlIGZ1bGwgc2h1dGRvd24K")
	if err != nil {
		return fmt.Sprintf("<an error occured while reading git rev: %v>", err)
	}
	if len(data) == 0 {
		return "<not available>"
	}
	return strings.TrimSpace(string(data))
}

