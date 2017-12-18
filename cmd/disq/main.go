package main

import (
	"flag"

	"io/ioutil"
	"os"

	"os/signal"
	"syscall"

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

	book, err := book.FromConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to compile config file")
	}

	s := disq.FromBook(book)
	s.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	run := true
	for run {
		select {
		case err = <-s.ErrorStream:
			log.WithError(err).Error("Error!")
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				// reload
			case syscall.SIGINT:
				log.Info("SIGINT")
				s.Stop()
				run = false
			case syscall.SIGTERM:
				log.Info("SIGTERM")
				s.Stop()
				run = false
			case syscall.SIGQUIT:
				log.Info("SIGQUIT")
				s.Stop()
				run = false
			default:
				log.Info("Unknown signal:", sig.String())
			}
		}
	}

	log.Info("Stopped.")
}
