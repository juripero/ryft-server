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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntervalAlignment(t *testing.T) {
	assert := assert.New(t)
	date := time.Date(2009, 4, 2, 3, 4, 5, 6, time.UTC)
	check := func(i string, expected time.Time) {
		proc := NewInterval(i)
		err := proc.Parse()
		assert.NoError(err)
		date := proc.Truncate(date)
		fmt.Printf("%s\n", date)
		assert.WithinDuration(expected, date, time.Millisecond)
	}
	check("year", time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC))
	check("1y", time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC))
	check("month", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("1M", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("2M", time.Date(2009, 3, 1, 0, 0, 0, 0, time.UTC))
	check("quarter", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("week", time.Date(2009, 3, 30, 0, 0, 0, 0, time.UTC))
	check("1w", time.Date(2009, 3, 30, 0, 0, 0, 0, time.UTC))
	check("2w", time.Date(2009, 3, 30, 0, 0, 0, 0, time.UTC))
	check("day", time.Date(2009, 4, 2, 0, 0, 0, 0, time.UTC))
	check("2d", time.Date(2009, 4, 2, 0, 0, 0, 0, time.UTC))
	check("100d", time.Date(2008, 12, 25, 0, 0, 0, 0, time.UTC))
}
