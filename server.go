package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const (
	serverOK = iota
	serverClosed
	serverFailed
)

type server struct {
	cmd     *exec.Cmd
	name    string
	jar     string
	java    string
	wd      string
	args    string
	log     string
	stop    string
	memMax  int
	memMin  int
	status  byte
	stopped bool
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
	if s.jar == "" {
		return errors.New("jar must be specified")
	}
	if s.java == "" {
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
	s.updateStop()
	s.status = 1
	return nil
}

func (s *server) updateStop() {
	if stopped {
		s.stopped = true
	}
	_, err := os.Open(s.stop)
	s.stopped = err == nil
}

func (s *server) start() (err error) {
	os.Remove(s.log)
	logFil, err := os.Create(s.log)
	if err != nil {
		s.status = serverFailed
		return
	}
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
	s.cmd.Dir = s.wd
	s.cmd.Stdout = logFil
	s.cmd.Stderr = logFil
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
	}()
	updateStatus <- struct{}{}
	return
}

func (s *server) stopCmd() {
	err := s.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		s.cmd.Process.Kill()
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
