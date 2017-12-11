package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
)

//go:generate bash geninfo.sh

func main() {
	log.Infof(" ** disq **")
	log.Infof("Build at: %s", color.MagentaString("%s", BuildAt()))
	log.Infof("Git Revision: \n%s", color.MagentaString("%s", DecodeGitRev()))
}
