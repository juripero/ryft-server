/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

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

func (o *Observer) Follow(name string, size int) (ch chan fsnotify.Op) {
	if size == 0 {
		ch = make(chan fsnotify.Op)
	} else {
		ch = make(chan fsnotify.Op, size)
	}

	o.c <- control{name: name, ch: ch}
	return ch
}

func (o *Observer) Unfollow(name string) {
	o.c <- control{name: name, ch: nil}
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
				go func() {
					ch <- e.Op
				}()
			}
		case err := <-o.w.Errors:
			log.Printf("PROC: error:%s", err.Error())
		}
	}
}
