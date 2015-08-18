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
	// if size == 0 {
	ch = make(chan fsnotify.Op)
	// } else {
	// 	ch = make(chan fsnotify.Op, size)
	// }

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
				o.m[c.name] = c.ch
			} else {
				delete(o.m, c.name)
			}
		case e := <-o.w.Events:
			if ch, ok := o.m[e.Name]; ok {
				ch <- e.Op
			}
		case err := <-o.w.Errors:
			log.Printf("fsobserver error:%s", err.Error())
		}
	}
}
