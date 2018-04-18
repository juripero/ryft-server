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

// test Index
func TestIndexSimple(t *testing.T) {
	idx := NewIndex("a.txt", 1, 2)
	assert.NotNil(t, idx)
	assert.Equal(t, "a.txt", idx.File)
	assert.EqualValues(t, 1, idx.Offset)
	assert.EqualValues(t, 2, idx.Length)
	assert.Empty(t, idx.Host)
	assert.Equal(t, `{a.txt#1, len:2, d:0}`, idx.String())

	idx.UpdateHost("localhost")
	assert.Equal(t, "localhost", idx.Host)

	idx.UpdateHost("ryft.com") // shouldn't be changed
	assert.Equal(t, "localhost", idx.Host)

	idx2 := NewIndexCopy(idx)
	//assert.False(t, idx == idx2)
	assert.EqualValues(t, idx.String(), idx2.String())

	idx.Release()
	assert.Empty(t, idx.File)
	assert.Empty(t, idx.Host)
}

// test CSV marshaling
func TestIndexMarshalCSV(t *testing.T) {
	idx := NewIndex("a.txt", 1, 2)
	data, err := idx.MarshalCSV()
	if assert.NoError(t, err) {
		assert.Equal(t, []string{
			"a.txt",
			"1",
			"2",
			"0",
			"",
		}, data)
	}
}

// TODO: test index pool in many goroutines

// test IndexFile
func TestIndexFile(t *testing.T) {
	f := NewIndexFile("\n", 0)
	f.Add(NewIndex("1.txt", 100, 50).SetDataPos(0))
	f.Add(NewIndex("2.txt", 200, 50).SetDataPos(51))
	assert.EqualValues(t, `delim:#0a, width:0, opt:0
{1.txt#100 [0..50)}
{2.txt#200 [51..101)}`, f.String())

	f.Clear()
	f.Width = 10
	assert.EqualValues(t, `delim:#0a, width:10, opt:0`, f.String())
	assert.EqualValues(t, 0, f.Len())

	f.Delim = "\n\f\n"
	f.Width = 0

	// 1.txt:
	// hello-00000
	// hello-11111
	// hello-22222
	f.Add(NewIndex("1.txt", 0, 11).SetDataPos(0))
	f.Add(NewIndex("1.txt", 11, 11).SetDataPos(14))
	f.Add(NewIndex("1.txt", 22, 11).SetDataPos(28))

	// 2.txt:
	// 33333-hello-33333
	// 44444-hello-44444
	// 55555-hello-55555
	f.Add(NewIndex("2.txt", 0, 17).SetDataPos(42))
	f.Add(NewIndex("2.txt", 17, 17).SetDataPos(62))
	f.Add(NewIndex("2.txt", 34, 17).SetDataPos(82))
	f.Add(NewIndex("2.txt", 51, 17).SetDataPos(102))

	// 3.txt:
	// 77777-hello
	// 88888-hello
	// 99999-hello
	f.Add(NewIndex("3.txt", 0, 11).SetDataPos(122))
	f.Add(NewIndex("3.txt", 11, 11).SetDataPos(136))
	f.Add(NewIndex("3.txt", 22, 11).SetDataPos(150))

	assert.EqualValues(t, `delim:#0a0c0a, width:0, opt:0
{1.txt#0 [0..11)}
{1.txt#11 [14..25)}
{1.txt#22 [28..39)}
{2.txt#0 [42..59)}
{2.txt#17 [62..79)}
{2.txt#34 [82..99)}
{2.txt#51 [102..119)}
{3.txt#0 [122..133)}
{3.txt#11 [136..147)}
{3.txt#22 [150..161)}`, f.String())

	tmp, shift, err := f.Unwind(NewIndex("0.dat", 1000, 5).SetFuzziness(1), 0)
	assert.EqualValues(t, `{0.dat#1000, len:5, d:1}`, tmp.String())
	assert.EqualValues(t, 0, shift)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no base found")
	}

	tmp, shift, err = f.Unwind(NewIndex("0.dat", 0, 5).UpdateHost("ryft.com"), 1)
	assert.EqualValues(t, `{1.txt#0, len:5, d:0}`, tmp.String())
	assert.EqualValues(t, "ryft.com", tmp.Host)
	assert.EqualValues(t, 0, shift)
	assert.NoError(t, err)

	tmp, shift, err = f.Unwind(NewIndex("0.dat", 14, 11), -1)
	assert.EqualValues(t, `{1.txt#11, len:11, d:0}`, tmp.String())
	assert.EqualValues(t, 0, shift)
	assert.NoError(t, err)

	tmp, shift, err = f.Unwind(NewIndex("0.dat", 15, 5), 0)
	assert.EqualValues(t, `{1.txt#12, len:5, d:0}`, tmp.String())
	assert.EqualValues(t, 0, shift)
	assert.NoError(t, err)

	tmp, shift, err = f.Unwind(NewIndex("0.dat", 15, 15), 0)
	assert.EqualValues(t, `{1.txt#12, len:10, d:0}`, tmp.String())
	assert.EqualValues(t, 0, shift)
	assert.NoError(t, err)

	tmp, shift, err = f.Unwind(NewIndex("0.dat", 10, 10), 5)
	assert.EqualValues(t, `{1.txt#11, len:6, d:0}`, tmp.String())
	assert.EqualValues(t, 4, shift)
	assert.NoError(t, err)

	tmp, shift, err = f.Unwind(NewIndex("0.dat", 10, 25), 5)
	assert.EqualValues(t, `{1.txt#11, len:11, d:0}`, tmp.String())
	assert.EqualValues(t, 4, shift)
	assert.NoError(t, err)
}
