package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func processConf(f *os.File) (err error) {
	rdr := bufio.NewReader(f)
	var line, lineTrim string
	var curServ *server
	var lineNum, ind int
	var key, value string
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
			case "log":
				os.Remove(value)
				var logFil *os.File
				logFil, err = os.Create(value)
				if err != nil {
					log.Println("Can't create log file")
					log.Println(err)
					log.Println("Continuing to log to stdout")
					err = nil
					continue
				}
				log.SetOutput(logFil)
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
	}
	if err == io.EOF {
		err = nil
	}
	return
}
