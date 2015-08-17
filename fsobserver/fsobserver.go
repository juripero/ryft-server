package fsobserver

import (
	"log"

	"github.com/go-fsnotify/fsnotify"
)

type Observer struct {
	w        *fsnotify.Watcher
	commands chan command
	paths    map[string]chan fsnotify.Op
}

func NewObserver(dir string) (o *Observer, err error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = w.Add(dir)
	if err != nil {
		return nil, err
	}

	o = &Observer{}
	o.commands = make(chan command, 256)
	o.paths = make(map[string]chan fsnotify.Op)
	o.w = w
	go o.commander()
	return o, nil
}

func (o *Observer) Stop() {
	o.commands <- stopCommand()
	o.w.Close()
}

func (o *Observer) Wait(file string) fsnotify.Op {
	c := waitCommand(file)
	o.commands <- c
	return <-c.eventc
}

func (o *Observer) WaitForCreate(file string) {
	for {
		if e := o.Wait(file); e&fsnotify.Create == fsnotify.Create {
			break
		}
	}
}

func (o *Observer) WaitForWrite(file string) {
	for {
		if e := o.Wait(file); e&fsnotify.Write == fsnotify.Write {
			break
		}
	}
}

func (o *Observer) commander() {
	for {
		select {
		case c := <-o.commands:
			switch c.cmd {
			case stop:
				return
			case wait:
				o.paths[c.path] = c.eventc
			}
		case event := <-o.w.Events:
			if eventc, ok := o.paths[event.Name]; ok {
				delete(o.paths, event.Name)
				eventc <- event.Op
				close(eventc)
			}
		case err := <-o.w.Errors:
			log.Printf("fsnotify wather error:%+v", err)
		}
	}
}
