package main

import (
	_ "embed"
	"flag"
	"log"
	"os"
	"path/filepath"
)

var (
	serv      []*server
	statusLoc string
	stopLoc   string
	watchConf bool
	stopped   bool
	rootMode  bool
	reset     chan struct{}
	stop      chan struct{}
)

//go:embed mcstarter.conf
var example []byte

func main() {
	if os.Getenv("USER") == "root" {
		rootMode = true
		log.Println("Starting in root mode...")
	} else {
		log.Println("Starting in user mode...")
	}
	flag.Parse()
	var confLoc string
	if len(flag.Args()) > 0 {
		confLoc = flag.Arg(0)
	} else {
		confLoc = "/etc/mcstarter.conf"
		_, err := os.Open("/etc/mcstarter.conf")
		if !rootMode && os.IsNotExist(err) {
			confLoc = "mcstarter.conf"
		}
	}
	if !filepath.IsAbs(confLoc) {
		wd, _ := os.Getwd()
		confLoc = filepath.Join(wd, confLoc)
	}
	startWatcher()
	statusLoop()
	reset = make(chan struct{})
	for {
		resetWatcher()
		log.SetOutput(os.Stdout)
		serv = make([]*server, 0)
		statusLoc = ""
		stopLoc = ""
		watchConf = true
		stopped = false
		stop = make(chan struct{})
		confFil, err := os.Open(confLoc)
		if os.IsNotExist(err) {
			confFil, err = os.Create(confLoc)
			if err != nil {
				log.Println("Conf file ("+confLoc+") doesn't exist and can't be created:", err)
				os.Exit(1)
			}
			_, err = confFil.Write(example)
			if err != nil {
				log.Println("Can't write example config...")
				log.Println(err)
				os.Exit(1)
			}
			wd, _ := os.Getwd()
			log.Println("Example config file created at", filepath.Join(wd, confFil.Name()))
			log.Println("Please configure your servers before restarting")
			os.Exit(0)
		}
		err = processConf(confFil)
		if err != nil {
			os.Exit(1)
		}
		if len(serv) == 0 {
			log.Println("No servers found in config. Exiting...")
			os.Exit(0)
		}
		if stopLoc != "" {
			err = addDirToWatcher(stopLoc, func() {
				stop <- struct{}{}
			})
			if err != nil {
				log.Println("can't watch stop file...")
				log.Println(err)
				os.Exit(1)
			}
		}
		if watchConf {
			err = addToWatcher(confFil.Name(), func() {
				reset <- struct{}{}
			})
			if err != nil {
				log.Println("can't watch config file...")
				log.Println(err)
				os.Exit(1)
			}
		}
		log.Println("Starting servers...")
		for _, s := range serv {
			err = addDirToWatcher(s.stop, s.stopOrStart)
			if err != nil {
				log.Println("can't watch", s.name, "stop file...")
				log.Println(err)
				os.Exit(1)
			}
			err = addDirToWatcher(s.input, s.processInput)
			if err != nil {
				log.Println("can't watch", s.name, "input file...")
				log.Println(err)
				os.Exit(1)
			}
			s.stopOrStart()
		}
		var shouldReset bool
		for !shouldReset {
			select {
			case <-reset:
				log.Println("config changed, stopping servers...")
				shouldReset = true
			case <-stop:
			}
			_, err = os.Open(stopLoc)
			if (err == nil && !stopped) || shouldReset {
				stopped = true
				for _, s := range serv {
					s.stopCmd()
				}
			} else if err != nil && stopped {
				stopped = false
				for _, s := range serv {
					s.stopOrStart()
				}
			}
		}
	}
}
