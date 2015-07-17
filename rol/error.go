package rol

import "strings"

type Error struct {
	s string
}

func (e *Error) Error() string {
	return e.s
}

const strangePattern = `SearchTreeNode: unable to execute HwController`

func (e *Error) IsStrangeError() bool {
	return strings.Contains(e.s, strangePattern)
}
