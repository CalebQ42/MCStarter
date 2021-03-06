package main

import (
	"log"
	"os"

	"github.com/CalebQ42/desktop/ini"
)

func processConf(f *os.File) (err error) {
	fil, err := ini.Parse(f)
	if err != nil {
		return
	}
	if fil.PreSection().HasKey("wd") {
		err = os.Chdir(fil.PreSection().Value("wd").String())
		if err != nil {
			log.Println("Can't change working directory")
			log.Println(err)
			log.Println("Exiting...")
			os.Exit(1)
		}
	}
	if fil.PreSection().HasKey("log") {
		os.Remove(fil.PreSection().Value("log").String())
		var logFil *os.File
		logFil, err = os.Create(fil.PreSection().Value("log").String())
		if err != nil {
			log.Println("Can't create a new log file")
			log.Println(err)
			log.Println("Exiting...")
			os.Exit(1)
		}
		log.SetOutput(logFil)
	}
	statusLoc = fil.PreSection().Value("status").String()
	stopLoc = fil.PreSection().Value("stop").String()
	watchConf = fil.PreSection().Value("watchConf").Bool()
	//TODO: process all servers.
}
