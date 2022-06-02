package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/fsnotify.v0"
)

var (
	watcher *fsnotify.Watcher
	watched map[string]func()
)

func startWatcher() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Println("Can't create file watcher")
		log.Println(err)
		os.Exit(1)
	}
	watched = make(map[string]func())
	go func() {
		var event *fsnotify.FileEvent
		for {
			event = <-watcher.Event
			f, ok := watched[filepath.Clean(event.Name)]
			if ok {
				f()
			}
		}
	}()
}

func addDirToWatcher(loc string, f func()) (err error) {
	for watched == nil {
		//since this is threaded, make sure that the map is created first
		time.Sleep(100 * time.Millisecond)
	}
	loc = filepath.Clean(loc)
	if _, ok := watched[loc]; ok {
		return nil
	}
	err = watcher.Watch(filepath.Dir(loc))
	if err != nil {
		return err
	}
	watched[loc] = f
	return nil
}

func addToWatcher(loc string, f func()) (err error) {
	for watched == nil {
		//since this is threaded, make sure that the map is created first
		time.Sleep(100 * time.Millisecond)
	}
	loc = filepath.Clean(loc)
	if _, ok := watched[loc]; ok {
		return nil
	}
	err = watcher.Watch(loc)
	if err != nil {
		return err
	}
	watched[loc] = f
	return nil
}

func resetWatcher() {
	for loc := range watched {
		watcher.RemoveWatch(loc)
		watcher.RemoveWatch(filepath.Dir(loc))
	}
}
