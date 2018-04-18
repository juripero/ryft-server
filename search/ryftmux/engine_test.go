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

package ryftmux

import (
	"fmt"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search/testfake"
	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// set test log level
func testSetLogLevel() {
	SetLogLevelString(testLogLevel)
	testfake.SetLogLevelString(testLogLevel)
}

// create new fake engine
func newFake(records, errors int) *testfake.Engine {
	engine, _ := testfake.NewEngine(fmt.Sprintf("/tmp/ryft-%u", time.Now().UnixNano()), "ryftmux")
	engine.SearchReportRecords = records
	engine.SearchReportErrors = errors
	return engine
}

// test engine options
func TestEngineOptions(t *testing.T) {
	testSetLogLevel()

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())

	backend := newFake(1, 0)
	engine, err := NewEngine(backend)
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, fmt.Sprintf("ryftmux{backends:[fake{home:%s/%s}]}", backend.MountPoint, backend.HomeDir), engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"index-host": "",
		}, engine.Options())
	}
}
