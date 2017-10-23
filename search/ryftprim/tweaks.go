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

package ryftprim

import (
	"fmt"
	"strings"
)

func NewTweakOpts(data map[string][]string) *TweaksOpts {
	return &TweaksOpts{data}
}

type TweaksOpts struct {
	data map[string][]string
}

func (t TweaksOpts) String() string {
	return fmt.Sprintf("%q", t.data)
}

func (t TweaksOpts) Data() map[string][]string {
	return t.data
}

func (t TweaksOpts) GetOptions(mode, backend, primitive string) []string {
	try := [][]string{
		[]string{mode, backend, primitive},
		[]string{backend, primitive},
		[]string{mode, primitive},
		[]string{mode, backend},
		[]string{primitive},
		[]string{backend},
		[]string{mode},
	}
	for _, el := range try {
		key := strings.Join(el, ".")
		if v, ok := t.data[key]; ok {
			return v
		}
	}
	return []string{}
}

func (t *TweaksOpts) SetOptions(value []string, mode, backend, primitive string) {
	keyStack := []string{}
	if mode != "" {
		keyStack = append(keyStack, mode)
	}
	if backend != "" {
		keyStack = append(keyStack, backend)
	}
	if primitive != "" {
		keyStack = append(keyStack, primitive)
	}
	t.data[strings.Join(keyStack, ".")] = value
}
