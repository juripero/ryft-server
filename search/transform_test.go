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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// test regexp-match transformation
func TestRegexpMatch(t *testing.T) {
	// good case
	check := func(expr string, in string, expectedOut string, expectedSkip bool) {
		tx, err := NewRegexpMatch(expr)
		if !assert.NoError(t, err) {
			return
		}

		out, skip, err := tx.Process([]byte(in))
		if assert.NotNil(t, out) && assert.NoError(t, err) {
			assert.EqualValues(t, expectedOut, out)
			assert.EqualValues(t, expectedSkip, skip)
		}
	}

	// bad case
	bad := func(expr string, in string, expectedError string) {
		tx, err := NewRegexpMatch(expr)
		if err != nil {
			assert.Contains(t, err.Error(), expectedError)
			return
		}

		out, _, err := tx.Process([]byte(in))
		if assert.Nil(t, out) && assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check("$apple^", "hello", "hello", true)
	check("^.*ell.*$", "hello", "hello", false)
	bad("$(", "hello", "error parsing regexp")
}

// test regexp-replace transformation
func TestRegexpReplace(t *testing.T) {
	// good case
	check := func(expr, templ string, in string, expectedOut string, expectedSkip bool) {
		tx, err := NewRegexpReplace(expr, templ)
		if !assert.NoError(t, err) {
			return
		}

		out, skip, err := tx.Process([]byte(in))
		if assert.NotNil(t, out) && assert.NoError(t, err) {
			assert.EqualValues(t, expectedOut, out)
			assert.EqualValues(t, expectedSkip, skip)
		}
	}

	// bad case
	bad := func(expr, templ string, in string, expectedError string) {
		tx, err := NewRegexpReplace(expr, templ)
		if err != nil {
			assert.Contains(t, err.Error(), expectedError)
			return
		}

		out, _, err := tx.Process([]byte(in))
		if assert.Nil(t, out) && assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check("$apple^", "$1", "hello", "hello", false) // keep as is
	check("^.*(ell).*$", "Z$1", "hello", "Zell", false)
	bad("$(", "$1", "hello", "error parsing regexp")
}

// test script-call transformation
func TestScriptCall(t *testing.T) {
	// good case
	check := func(script []string, in string, expectedOut string, expectedSkip bool) {
		tx, err := NewScriptCall(script, "", script[0], script[1:])
		if !assert.NoError(t, err) {
			return
		}

		out, skip, err := tx.Process([]byte(in))
		if assert.NotNil(t, out) && assert.NoError(t, err) {
			assert.EqualValues(t, expectedOut, out)
			assert.EqualValues(t, expectedSkip, skip)
		}
	}

	// bad case
	bad := func(script []string, in string, expectedError string) {
		tx, err := NewScriptCall(script, "", "", nil)
		if err != nil {
			assert.Contains(t, err.Error(), expectedError)
			return
		}

		out, _, err := tx.Process([]byte(in))
		if assert.Nil(t, out) && assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check([]string{"/bin/cat"}, "hello", "hello", false)
	check([]string{"/bin/grep", "hell"}, "apple\nhello\norange\n", "hello\n", false)
	check([]string{"/bin/false"}, "hello", "hello", true)
	bad([]string{}, "hello", "no script path provided")
	bad([]string{"/bin/missing-script"}, "hello", "no valid script found")

	if tx, err := NewScriptCall([]string{"/bin/cat"}, "", "cat", nil); assert.NoError(t, err) {
		assert.EqualValues(t, "script(cat)", tx.String())
	}
	if tx, err := NewScriptCall([]string{"/bin/cat", "-", "a"}, "", "cat", []string{"-", "a"}); assert.NoError(t, err) {
		assert.EqualValues(t, "script(cat,-,a)", tx.String())
	}
}
