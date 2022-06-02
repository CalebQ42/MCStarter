package main

import (
	"log"
	"os"
)

var (
	updateStatus chan struct{}
)

func statusLoop() {
	updateStatus = make(chan struct{})
	var stat *os.File
	var err error
	go func() {
		for {
			<-updateStatus
			if statusLoc == "" {
				continue
			}
			os.Remove(statusLoc)
			stat, err = os.Create(statusLoc)
			if err != nil {
				log.Println("Can't update status file:")
				log.Println(err)
				continue
			}
			for _, s := range serv {
				_, err = stat.WriteString(s.name + ": " + s.statusString() + "\n")
				if err != nil {
					log.Println("Can't update status file:")
					log.Println(err)
				}
			}
		}
	}()
}
