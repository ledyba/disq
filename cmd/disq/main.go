package main

import (
	"flag"

	"io/ioutil"
	"os"

	"syscall"

	"os/signal"

	"sync"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/ledyba/disq"
	"github.com/ledyba/disq/book"
	"github.com/ledyba/disq/conf"
	"github.com/ledyba/disq/zabbix"
)

//go:generate bash geninfo.sh

var config = flag.String("config", "./config.json", "Config file path")
var zabbixHost = flag.String("zabbix", "", "Zabbix server addr")
var verbose = flag.Bool("v", false, "BE VERBOSE.")

var hostname string
var sender *zabbix.Sender

func reload(s *disq.Server) {
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
		log.WithField("Module", "Reload").WithError(err).Error("Failed to load config file")
		return
	}

	cfg, err := conf.Load(dat)
	if err != nil {
		log.WithField("Module", "Reload").WithError(err).Error("Failed to parse config file")
		return
	}

	b, err := book.FromConfig(cfg)
	if err != nil {
		log.WithField("Module", "Reload").WithError(err).Error("Failed to compile config file")
		return
	}

	err = s.Reload(b)
	if err != nil {
		log.WithField("Module", "Reload").WithError(err).Error("Failed to reload book")
	}
}

func sendZabbix(msg string) {
	if sender == nil {
		return
	}
	log.WithField("Module", "Zabbix").Debugf("Sending: %s", msg)
	payload := []*zabbix.Metric{
		zabbix.NewMetric(hostname, "disq.errors", msg),
	}
	resp, err := sender.Send(zabbix.NewPacket(payload))
	if err != nil {
		log.WithField("Module", "Zabbix").Errorf("Error while sending: %v", err)
	} else {
		log.WithField("Module", "Zabbix").Debugf("Sent successfully: %s", string(resp))
	}
}

func main() {
	var err error
	flag.Parse()
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof(`
***** disq *****
Hostname: %s
Build at: %s"
Git Revision:
%s
****************`,
		color.MagentaString("%s", hostname),
		color.MagentaString("%s", buildAt()),
		color.MagentaString("%s", gitRev()))

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	if len(*zabbixHost) > 0 {
		sender = zabbix.NewSender(*zabbixHost)
	}

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

	var wg sync.WaitGroup

	errorHandleDone := make(chan struct{})

	wg.Add(1)
	go func() {
		log.WithField("Module", "ErrorHandler").Info("started.")
		defer wg.Done()
		defer log.WithField("Module", "ErrorHandler").Info("shutdown succeeded.")
		for {
			select {
			case err := <-s.ErrorStream:
				log.WithField("Module", "ErrorHandler").WithError(err).Error("Error!")
				sendZabbix(fmt.Sprintf("[Error]%s", err.Error()))
			case <-errorHandleDone:
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		log.WithField("Module", "SignalHandler").Info("started")
		defer wg.Done()
		defer log.WithField("Module", "SignalHandler").Info("shutdown succeeded.")
		for {
			select {
			case sig := <-sigChan:
				switch sig {
				case syscall.SIGHUP:
					log.Info("SIGNAL: SIGHUP")
					reload(s)
				case syscall.SIGINT:
					log.Info("SIGNAL: SIGINT")
					s.Stop()
					errorHandleDone <- struct{}{}
					return
				case syscall.SIGTERM:
					log.Info("SIGNAL: SIGTERM")
					s.Stop()
					errorHandleDone <- struct{}{}
					return
				case syscall.SIGQUIT:
					log.Info("SIGNAL: SIGQUIT")
					s.Stop()
					errorHandleDone <- struct{}{}
					return
				default:
					log.Info("SIGNAL: Unknown signal:", sig.String())
				}
			}
		}
	}()
	log.Info("All subsystems started.")
	sendZabbix("Started")

	wg.Wait()

	log.Info("All subsystems stopped.")
}
