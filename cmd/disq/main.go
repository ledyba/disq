package main

import (
	"flag"

	"io/ioutil"
	"os"

	"syscall"

	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/ledyba/disq"
	"github.com/ledyba/disq/book"
	"github.com/ledyba/disq/conf"
)

//go:generate bash geninfo.sh

var config = flag.String("config", "./config.json", "Config file path")

func main() {
	flag.Parse()

	var err error
	log.Infof("***** disq *****")
	log.Infof("Build at: %s", color.MagentaString("%s", buildAt()))
	log.Infof("Git Revision: \n%s", color.MagentaString("%s", gitRev()))
	log.Infof("****************")

	dat, err := func() ([]byte, error) {
		var err error
		f, err := os.Open(*config)
		if err != nil {
			return nil, err
		}
		dat, err := ioutil.ReadAll(f)
		return dat, err
	}()

	if err != nil {
		log.WithError(err).Fatal("Failed to load config file")
	}

	cfg, err := conf.Load(dat)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse config file")
	}

	b, err := book.FromConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to compile config file")
	}

	s := disq.FromBook(b)
	s.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	run := true
	log.Info("All subsystems started.")
	for run {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Info("SIGNAL: SIGHUP")
				// reload
			case syscall.SIGINT:
				log.Info("SIGNAL: SIGINT")
				s.Stop()
				run = false
			case syscall.SIGTERM:
				log.Info("SIGNAL: SIGTERM")
				s.Stop()
				run = false
			case syscall.SIGQUIT:
				log.Info("SIGNAL: SIGQUIT")
				s.Stop()
				run = false
			default:
				log.Info("SIGNAL: Unknown signal:", sig.String())
			}
		case err = <-s.ErrorStream:
			log.WithError(err).Error("Error!")
		}
	}

	log.Info("All subsystems stopped.")
}
