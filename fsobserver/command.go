package fsobserver

import "github.com/go-fsnotify/fsnotify"

type cmdT int

const (
	stop cmdT = 0
	wait cmdT = 1
)

type command struct {
	cmd    cmdT
	path   string
	eventc chan fsnotify.Op
}

func stopCommand() command {
	return command{cmd: stop}
}
func waitCommand(path string) command {
	return command{wait, path, make(chan fsnotify.Op, 1)}
}
