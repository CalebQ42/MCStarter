package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CalebQ42/desktop/ini"
)

const (
	serverOK = iota
	serverClosed
	serverFailed
)

type server struct {
	cmd      *exec.Cmd
	cmdInput io.Writer
	script   string
	name     string
	jar      string
	java     string
	wd       string
	args     string
	log      string
	stop     string
	input    string
	memMax   int
	memMin   int
	status   byte
	stopped  bool
}

func newServer(name string, sec *ini.Section) (s *server, err error) {
	s = new(server)
	s.name = name
	s.script = sec.Value("script").String()
	s.jar = sec.Value("jar").String()
	s.java = sec.Value("java").String()
	s.wd = sec.Value("wd").String()
	s.memMax = sec.Value("memMax").Int()
	s.memMin = sec.Value("memMin").Int()
	s.args = sec.Value("args").String()
	s.log = sec.Value("log").String()
	s.stop = sec.Value("stop").String()
	s.input = sec.Value("input").String()
	return s, s.validate()
}

func (s server) statusString() string {
	switch s.status {
	case 0:
		return "OK"
	case 1:
		if s.stopped {
			return "Stopped"
		}
		return "Closed"
	case 2:
		if s.stopped {
			return "Stopped"
		}
		return "Failed"
	}
	return "Unknown"
}

func (s *server) validate() error {
	if s.jar == "" && s.script == "" {
		return errors.New("jar or script must be specified")
	}
	if s.java == "" && s.jar != "" {
		s.java = "java"
	}
	if s.wd == "" {
		s.wd = s.name
	}
	if s.log == "" {
		s.log = "log"
	}
	if !filepath.IsAbs(s.log) {
		s.log = filepath.Join(s.wd, s.log)
	}
	if s.stop == "" {
		s.stop = "stop"
	}
	if !filepath.IsAbs(s.stop) {
		s.stop = filepath.Join(s.wd, s.stop)
	}
	if s.input == "" {
		s.input = "input"
	}
	if !filepath.IsAbs(s.input) {
		s.input = filepath.Join(s.wd, s.input)
	}
	s.updateStop()
	s.status = 1
	return nil
}

func (s *server) updateStop() {
	if stopped {
		s.stopped = true
	}
	_, err := os.Open(s.stop)
	s.stopped = (err == nil)
}

func (s *server) start() (err error) {
	os.Remove(s.log)
	logFil, err := os.Create(s.log)
	if err != nil {
		s.status = serverFailed
		return
	}
	if s.jar != "" {
		args := make([]string, 0)
		if s.memMax != 0 {
			args = append(args, "-Xmx"+strconv.Itoa(s.memMax)+"M")
		}
		if s.memMin != 0 {
			args = append(args, "-Xms"+strconv.Itoa(s.memMin)+"M")
		}
		args = append(args, "-jar", s.jar)
		args = append(args, s.args)
		s.cmd = exec.Command(s.java, args...)
	} else {
		s.cmd = exec.Command(s.script)
	}
	var cmdInRdr io.Reader
	cmdInRdr, s.cmdInput = io.Pipe()
	s.cmd.Dir = s.wd
	s.cmd.Stdout = logFil
	s.cmd.Stderr = logFil
	s.cmd.Stdin = cmdInRdr
	log.Println(s.name, "started")
	err = s.cmd.Start()
	if err != nil {
		s.status = serverFailed
		s.cmd = nil
		return
	}
	s.status = serverOK
	go func() {
		s.cmd.Wait()
		if s.cmd.ProcessState.ExitCode() == 0 {
			s.status = serverClosed
		} else {
			s.status = serverFailed
		}
		s.cmd = nil
		updateStatus <- struct{}{}
		log.Println(s.name, "closed")
	}()
	updateStatus <- struct{}{}
	return
}

func (s *server) stopCmd() {
	if s.cmd != nil {
		err := s.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			s.cmd.Process.Kill()
		}
		s.cmd.Wait()
	}
}

func (s *server) processInput() {
	in, err := os.Open(s.input)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Println("Can't read input file for", s.name)
		log.Println(err)
		return
	}
	buf := bufio.NewReader(in)
	var line string
	empty := true
	for {
		line, err = buf.ReadString('\n')
		if line != "" && err != nil {
			err = nil
		} else if line == "" && err != nil {
			break
		}
		empty = false
		if !strings.HasSuffix(line, "\n") {
			line += "\n"
		}
		log.Println("Sending command \""+line[:len(line)-1]+"\" to", s.name)
		_, err = s.cmdInput.Write([]byte(line))
		if err != nil {
			log.Println("Can't send command:", line)
			log.Println(err)
		}
	}
	if !empty {
		os.Remove(s.input)
	}
}

func (s *server) stopOrStart() {
	s.updateStop()
	if s.cmd == nil && !s.stopped {
		err := s.start()
		if err != nil {
			log.Println("Can't start", s.name)
			log.Println(err)
			log.Println("Create then delete stop file to restart")
			return
		}
	} else if s.cmd != nil && s.stopped {
		s.stopCmd()
	}
}
