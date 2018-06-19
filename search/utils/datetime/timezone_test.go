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

package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadTimezone(t *testing.T) {
	assert := assert.New(t)
	success := func(name string, expected string) {
		loc, err := LoadTimezone(name)
		assert.NoError(err)
		moment := time.Date(2007, 1, 1, 12, 0, 0, 0, time.UTC)
		assert.Equal(expected, moment.In(loc).Format("Z07:00"))
	}

	fail := func(name string) {
		_, err := LoadTimezone(name)
		assert.Error(err)
	}

	success("", "Z")
	success("UTC", "Z")
	success("America/Chicago", "-06:00")
	success("Asia/Pontianak", "+07:00")
	success("08:00", "+08:00")
	success("-08:00", "-08:00")
	success("+01", "+01:00")
	success("-01", "-01:00")
	success("0200", "+02:00")
	success("-0200", "-02:00")
	success("+10:30", "+10:30")
	success("+04:51", "+04:51")

	fail("Russia/Cheboksary")
	fail("1")
	fail("+1")
	fail("-1")
	fail("-x1")
	fail("-:")
	fail("-:0")
	fail("-:01")
}
