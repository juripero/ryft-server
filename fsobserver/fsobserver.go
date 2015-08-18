package fsobserver

import (
	"log"

	"github.com/go-fsnotify/fsnotify"
)

type Observer struct {
	w *fsnotify.Watcher
	m map[string]chan fsnotify.Op
	c chan control
}

type control struct {
	name string
	ch   chan fsnotify.Op
}

func NewObserver(dir string) (o *Observer, err error) {
	var w *fsnotify.Watcher
	if w, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}

	if err = w.Add(dir); err != nil {
		return nil, err
	}

	o = &Observer{}
	o.m = make(map[string]chan fsnotify.Op)
	o.c = make(chan control, 256)
	o.w = w

	go o.process()

	return o, nil
}

func (o *Observer) Follow(name string) (ch chan fsnotify.Op) {
	ch = make(chan fsnotify.Op)
	o.c <- control{name: name, ch: ch}
	return ch
}

func (o *Observer) Unfollow(name string) {
	log.Printf("Unfollow: start %s", name)
	o.c <- control{name: name, ch: nil}
	log.Printf("Unfollow: end %s", name)
}

func (o *Observer) process() {
	for {
		select {
		case c := <-o.c:
			if c.ch != nil {
				log.Printf("PROC: adding %s ...", c.name)
				o.m[c.name] = c.ch
				log.Printf("PROC: add %s", c.name)
			} else {
				log.Printf("PROC: deleting %s ...", c.name)
				delete(o.m, c.name)
				log.Printf("PROC: del %s", c.name)
			}
		case e := <-o.w.Events:
			log.Printf("PROC: RAW %s", e)
			if ch, ok := o.m[e.Name]; ok {
				log.Printf("PROC: sending... %s", e)
				ch <- e.Op
				log.Printf("PROC: sent %s", e)
			}
		case err := <-o.w.Errors:
			log.Printf("PROC: error:%s", err.Error())
		}
	}
}
