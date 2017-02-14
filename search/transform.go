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

package search

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

// Transform is an abstract transformation rule
type Transform interface {
	Process(in []byte) (out []byte, skip bool, err error)
}

// regexp-match transformation
type regexpMatch struct {
	re *regexp.Regexp
}

// NewRegexpMatch creates new "regexp-match" transformation.
// is based on regexp.Regexp.Match function
func NewRegexpMatch(expr string) (*regexpMatch, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regexp-match expression: %s", err)
	}

	return &regexpMatch{re: re}, nil // OK
}

// do regexp-match transformation
func (t *regexpMatch) Process(in []byte) ([]byte, bool, error) {
	return in, !t.re.Match(in), nil
}

// get string expression
func (t *regexpMatch) String() string {
	return fmt.Sprintf(`match("%s")`, t.re.String())
}

// regexp-replace transformation
type regexpReplace struct {
	re       *regexp.Regexp
	template []byte // replace with
}

// NewRegexpReplace creates new "regexp-replace" transformation.
// is based on regexp.Regexp.ReplaceAll function
func NewRegexpReplace(expr string, template string) (*regexpReplace, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regexp-replace expression: %s", err)
	}

	return &regexpReplace{re: re, template: []byte(template)}, nil // OK
}

// do regexp-replace transformation
func (t *regexpReplace) Process(in []byte) ([]byte, bool, error) {
	return t.re.ReplaceAll(in, t.template), false, nil
}

// get string expression
func (t *regexpReplace) String() string {
	return fmt.Sprintf(`replace("%s","%s")`, t.re.String(), t.template)
}

// script-call transformation
type scriptCall struct {
	path []string // path + args
	wdir string   // working directory
}

// NewScriptCall created new "script-call" transformation.
func NewScriptCall(pathAndArgs []string, workDir string) (*scriptCall, error) {
	if len(pathAndArgs) == 0 {
		return nil, fmt.Errorf("no script path provided")
	}

	if _, err := os.Stat(pathAndArgs[0]); err != nil {
		return nil, fmt.Errorf("no valid script found: %s", err)
	}
	// TODO: check script is executable

	return &scriptCall{path: pathAndArgs, wdir: workDir}, nil // OK
}

// do script-call transformation
func (t *scriptCall) Process(in []byte) ([]byte, bool, error) {
	// create command
	cmd := exec.Command(t.path[0], t.path[1:]...)
	cmd.Dir = t.wdir

	cmd.Stdin = bytes.NewReader(in)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return in, true, nil // skipped
		}
	}

	return out, false, err
}

// get string expression
func (t *scriptCall) String() string {
	return fmt.Sprintf(`script(%s)`, t.path)
}
