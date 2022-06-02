package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func processConf(f *os.File) (err error) {
	rdr := bufio.NewReader(f)
	var line, lineTrim string
	var curServ *server
	var lineNum, ind int
	var key, value string
	var wdSet bool
	var tmpLog string
	for err == nil {
		line, err = rdr.ReadString('\n')
		if line != "" && err != nil {
			err = nil
		} else if err != nil {
			break
		}
		lineNum++
		lineTrim = strings.TrimSpace(line)
		if strings.HasPrefix(lineTrim, "#") || lineTrim == "" {
			continue
		}
		line = strings.TrimSuffix(line, "\n")
		lineTrim = strings.TrimSuffix(lineTrim, "\n")
		if strings.HasPrefix(lineTrim, "[") && strings.HasSuffix(lineTrim, "]") {
			if curServ != nil {
				err = curServ.validate()
				if err != nil {
					log.Println("Can't validate", curServ.name)
					log.Println(err)
					return
				}
				serv = append(serv, curServ)
			} else {
				if rootMode && !wdSet {
					log.Println("Working directory not specified, and run with root. Exiting...")
					os.Exit(1)
				}
				if tmpLog != "" {
					os.Remove(tmpLog)
					var logFil *os.File
					logFil, err = os.Create(tmpLog)
					if err != nil {
						log.Println("Can't create a new log file")
						log.Println(err)
						log.Println("Exiting...")
						os.Exit(1)
					}
					log.SetOutput(logFil)
				}
			}
			curServ = new(server)
			curServ.name = strings.Trim(lineTrim, "[]")
			if curServ.name == "" {
				log.Println("No name given for server at line", lineNum)
				err = errors.New("no name")
				return
			}
			continue
		}
		ind = strings.Index(line, "=")
		if ind == -1 {
			log.Println("Line", lineNum, "is not valid. Ignoring...")
			continue
		}
		key, value = line[:ind], line[ind+1:]
		if curServ == nil {
			switch key {
			case "wd":
				err = os.Chdir(value)
				if err != nil {
					log.Println("Can't change working directory")
					log.Println(err)
					log.Println("Exiting...")
					os.Exit(1)
				}
				wdSet = true
				if tmpLog != "" {
					os.Remove(tmpLog)
					var logFil *os.File
					logFil, err = os.Create(tmpLog)
					if err != nil {
						log.Println("Can't create a new log file")
						log.Println(err)
						log.Println("Exiting...")
						os.Exit(1)
					}
					log.SetOutput(logFil)
				}
			case "log":
				if wdSet || filepath.IsAbs(value) {
					os.Remove(value)
					var logFil *os.File
					logFil, err = os.Create(value)
					if err != nil {
						log.Println("Can't create a new log file")
						log.Println(err)
						log.Println("Exiting...")
						os.Exit(1)
					}
					log.SetOutput(logFil)
				} else {
					tmpLog = value
				}
			case "status":
				statusLoc = value
			case "stop":
				stopLoc = value
			case "watchConf":
				if value == "false" {
					watchConf = false
				}
			default:
				log.Println("Invalid key at line", lineNum, ":", key, "ignoring...")
			}
			continue
		}
		switch key {
		case "jar":
			curServ.jar = value
		case "java":
			curServ.java = value
		case "wd":
			curServ.wd = value
		case "memMax":
			var val int
			val, err = strconv.Atoi(value)
			if err != nil {
				log.Println("memMax at line", lineNum, "not a number. ignoring...")
				err = nil
				continue
			}
			curServ.memMax = val
		case "memMin":
			var val int
			val, err = strconv.Atoi(value)
			if err != nil {
				log.Println("memMin at line", lineNum, "not a number. ignoring...")
				err = nil
				continue
			}
			curServ.memMin = val
		case "args":
			curServ.args = value
		case "log":
			curServ.log = value
		case "stop":
			curServ.stop = value
		default:
			log.Println("Invalid key at line", lineNum, ":", key, "ignoring...")
		}
	}
	if curServ != nil {
		err = curServ.validate()
		if err != nil {
			log.Println("Can't validate", curServ.name)
			log.Println(err)
			return
		}
		serv = append(serv, curServ)
	} else {
		if rootMode && !wdSet {
			log.Println("Working directory not specified, and run with root. Exiting...")
			os.Exit(1)
		}
		if tmpLog != "" {
			os.Remove(tmpLog)
			var logFil *os.File
			logFil, err = os.Create(tmpLog)
			if err != nil {
				log.Println("Can't create a new log file")
				log.Println(err)
				log.Println("Exiting...")
				os.Exit(1)
			}
			log.SetOutput(logFil)
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}
