package main

import (
	_ "embed"
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
		confFil, err := os.Open("/etc/mcstarter.conf")
		if err != nil && !rootMode {
			confFil, err = os.Open("mcstarter.conf")
			if err != nil {
				confFil, err = os.Create("mcstarter.conf")
				if err != nil {
					log.Println("Can't find /etc/mcstarter.conf, mcstarter.conf, and can't create an example file...")
					log.Println(err)
					os.Exit(1)
				}
				_, err = confFil.Write(example)
				if err != nil {
					log.Println("Can't write to example config...")
					log.Println(err)
					os.Exit(1)
				}
				wd, _ := os.Getwd()
				log.Println("Example config file created at", filepath.Join(wd, confFil.Name()))
				log.Println("Please configure your servers before restarting")
				os.Exit(0)
			}
		} else if err != nil {
			confFil, err = os.Create("/etc/mcstarter.conf")
			if err != nil {
				log.Println("Can't find /etc/mcstarter.conf, mcstarter.conf, and can't create an example file...")
				log.Println(err)
				os.Exit(1)
			}
			_, err = confFil.Write(example)
			if err != nil {
				log.Println("Can't write to example config...")
				log.Println(err)
				os.Exit(1)
			}
			log.Println("Example config file created at /etc/mcstarter.conf")
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
		for _, s := range serv {
			err = addDirToWatcher(s.stop, s.stopOrStart)
			if err != nil {
				log.Println("can't watch", s.name, "stop file...")
				log.Println(err)
				os.Exit(1)
			}
			s.stopOrStart()
		}
		log.Println("Starting servers...")
		var shouldReset bool
		for !shouldReset {
			select {
			case <-reset:
				log.Println("config changed, restarting...")
				shouldReset = true
			case <-stop:
			}
			_, err = os.Open(stopLoc)
			if (err == nil && !stopped) || shouldReset {
				stopped = true
				for _, s := range serv {
					s.stopOrStart()
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
