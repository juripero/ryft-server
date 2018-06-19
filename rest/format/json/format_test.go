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

package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test format options
func TestFormatOptions(t *testing.T) {
	// bad option type
	_, err := New(map[string]interface{}{
		"fields": 555,
	})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unsupported option type")
	}

	// fields from string
	fmt1, err := New(map[string]interface{}{
		"fields": "a,b",
	})
	if assert.NoError(t, err) && assert.NotNil(t, fmt1) {
		assert.EqualValues(t, fmt1.Fields, []string{"a", "b"})
	}

	// fields from []string
	fmt2, err := New(map[string]interface{}{
		"fields": []string{"a", "b"},
	})
	if assert.NoError(t, err) && assert.NotNil(t, fmt2) {
		assert.EqualValues(t, fmt2.Fields, []string{"a", "b"})
	}

	// AddFields
	fmt2.AddFields("c,d")
	assert.EqualValues(t, fmt2.Fields, []string{"a", "b", "c", "d"})
}
