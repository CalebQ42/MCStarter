package main

import (
	_ "embed"
	"log"
	"os"
)

var (
	serv      []*server
	statusLoc string
	stopLoc   string
	watchConf bool
	stopped   bool
	reset     chan struct{}
	stop      chan struct{}
)

//go:embed mcstarter.conf
var example []byte

func main() {
	startWatcher()
	statusLoop()
	reset = make(chan struct{})
	for {
		resetWatcher()
		statusLoc = ""
		serv = make([]*server, 0)
		log.SetOutput(os.Stdout)
		stop = make(chan struct{})

		confFil, err := os.Open("/etc/mcstarter.conf")
		if err != nil {
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
					log.Println("err")
					os.Exit(1)
				}
			}
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
			err = addToWatcher(stopLoc, func() {
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
			err = addToWatcher(s.stop, s.stopOrStart)
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
			select {
			case <-reset:
				shouldReset = true
			case <-stop:
			}
		}
	}
}
