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

package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Convert to JSON
func TestQueryToJSON(t *testing.T) {
	// check function
	check := func(hasRec bool, query string, expected string) {
		q, err := ParseQueryOptEx(query, DefaultOptions(), IN_JRECORD, nil)
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, fmt.Sprintf("%+v", q))
			assert.EqualValues(t, hasRec, q.IsSomeStructured())
		}
	}

	check(false, `RAW_TEXT CONTAINS "hello"`, `(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(true, `JRECORD CONTAINS "hello"`, `(JRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `XRECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `CRECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD CONTAINS "hello"`, `(JRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.body CONTAINS "hello"`, `(JRECORD.body CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.[] CONTAINS "hello"`, `(JRECORD.[] CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD.[] CONTAINS "hello" AND RAW_TEXT CONTAINS "world"`,
		`AND{(JRECORD.[] CONTAINS EXACT("hello"))[es], (RAW_TEXT CONTAINS EXACT("world"))[es]}`)
	check(true, `RAW_TEXT CONTAINS "world" OR RECORD.[] CONTAINS "hello"`,
		`OR{(RAW_TEXT CONTAINS EXACT("world"))[es], (JRECORD.[] CONTAINS EXACT("hello"))[es]}`)
}

// Convert to XML
func TestQueryToXML(t *testing.T) {
	// check function
	check := func(hasRec bool, query string, expected string) {
		q, err := ParseQueryOptEx(query, DefaultOptions(), IN_XRECORD, nil)
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, fmt.Sprintf("%+v", q))
			assert.EqualValues(t, hasRec, q.IsSomeStructured())
		}
	}

	check(false, `RAW_TEXT CONTAINS "hello"`, `(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(true, `JRECORD CONTAINS "hello"`, `(JRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `XRECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `CRECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.body CONTAINS "hello"`, `(XRECORD.body CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.[] CONTAINS "hello"`, `(XRECORD.[] CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD.[] CONTAINS "hello" AND RAW_TEXT CONTAINS "world"`,
		`AND{(XRECORD.[] CONTAINS EXACT("hello"))[es], (RAW_TEXT CONTAINS EXACT("world"))[es]}`)
	check(true, `RAW_TEXT CONTAINS "world" OR RECORD.[] CONTAINS "hello"`,
		`OR{(RAW_TEXT CONTAINS EXACT("world"))[es], (XRECORD.[] CONTAINS EXACT("hello"))[es]}`)
}

// Convert to CSV
func TestQueryToCSV(t *testing.T) {
	// check function
	check := func(hasRec bool, query string, expected string) {
		q, err := ParseQueryOptEx(query, DefaultOptions(), IN_CRECORD, nil)
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, fmt.Sprintf("%+v", q))
			assert.EqualValues(t, hasRec, q.IsSomeStructured())
		}
	}

	check(false, `RAW_TEXT CONTAINS "hello"`, `(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(true, `JRECORD CONTAINS "hello"`, `(JRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `XRECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `CRECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.body CONTAINS "hello"`, `(CRECORD.body CONTAINS EXACT("hello"))[es]`)
	check(true, `RECORD.[] CONTAINS "hello"`, `(CRECORD.[] CONTAINS EXACT("hello"))[es]`)

	check(true, `RECORD.[] CONTAINS "hello" AND RAW_TEXT CONTAINS "world"`,
		`AND{(CRECORD.[] CONTAINS EXACT("hello"))[es], (RAW_TEXT CONTAINS EXACT("world"))[es]}`)
	check(true, `RAW_TEXT CONTAINS "world" OR RECORD.[] CONTAINS "hello"`,
		`OR{(RAW_TEXT CONTAINS EXACT("world"))[es], (CRECORD.[] CONTAINS EXACT("hello"))[es]}`)
}

// Convert to CSV with column names
func TestQueryToCSV2(t *testing.T) {
	// check function
	check := func(hasRec bool, newFields map[string]string, query string, expected string) {
		q, err := ParseQueryOptEx(query, DefaultOptions(), IN_CRECORD, newFields)
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, fmt.Sprintf("%+v", q))
			assert.EqualValues(t, hasRec, q.IsSomeStructured())
		}
	}

	check(false, nil, `RAW_TEXT CONTAINS "hello"`, `(RAW_TEXT CONTAINS EXACT("hello"))[es]`)
	check(true, nil, `JRECORD CONTAINS "hello"`, `(JRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, nil, `XRECORD CONTAINS "hello"`, `(XRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, nil, `CRECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)

	check(true, nil, `RECORD CONTAINS "hello"`, `(CRECORD CONTAINS EXACT("hello"))[es]`)
	check(true, nil, `RECORD.body CONTAINS "hello"`, `(CRECORD.body CONTAINS EXACT("hello"))[es]`)
	check(true, nil, `RECORD.[] CONTAINS "hello"`, `(CRECORD.[] CONTAINS EXACT("hello"))[es]`)

	check(true, map[string]string{"body": "123"},
		`RECORD.body CONTAINS "hello"`,
		`(CRECORD.123 CONTAINS EXACT("hello"))[es]`)
	check(true, map[string]string{"body": "123"},
		`RECORD."body" CONTAINS "hello"`,
		`(CRECORD.123 CONTAINS EXACT("hello"))[es]`)

	check(true, map[string]string{"foo": "123"}, `RECORD.[]."foo" CONTAINS "hello" AND RAW_TEXT CONTAINS "world"`,
		`AND{(CRECORD.[].123 CONTAINS EXACT("hello"))[es], (RAW_TEXT CONTAINS EXACT("world"))[es]}`)
	check(true, map[string]string{"foo": "321"}, `RAW_TEXT CONTAINS "world" OR RECORD.[].foo."foo".123 CONTAINS "hello"`,
		`OR{(RAW_TEXT CONTAINS EXACT("world"))[es], (CRECORD.[].321.321.123 CONTAINS EXACT("hello"))[es]}`)
}
