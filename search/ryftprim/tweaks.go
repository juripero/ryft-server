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
	"unicode"

	"github.com/getryft/ryft-server/search/utils"
)

// Tweaks custom backend options and routing table
type Tweaks struct {
	Options map[string][]string // custom options

	// routing table: [primitive] => backend
	Router map[string]string
}

// ParseTweaks parses tweaks from engine options
func ParseTweaks(opts_ map[string]interface{}) (*Tweaks, error) {
	t := new(Tweaks)
	t.Options = make(map[string][]string)
	t.Router = make(map[string]string)

	if true { // [backward compatibility]
		// default options for all engines
		if v, ok := opts_["ryft-all-opts"]; ok {
			if vv, err := utils.AsStringSlice(v); err != nil {
				return nil, fmt.Errorf(`failed to parse "ryft-all-opts" option: %s`, err)
			} else {
				t.SetOptions("default", "", "", vv)
			}
		}

		// `ryftprim` options
		if v, ok := opts_["ryftprim-opts"]; ok {
			if vv, err := utils.AsStringSlice(v); err != nil {
				return nil, fmt.Errorf(`failed to parse "ryftprim-opts" option: %s`, err)
			} else {
				t.SetOptions("", "ryftprim", "", vv)
			}
		}

		// `ryftx` options
		if v, ok := opts_["ryftx-opts"]; ok {
			if vv, err := utils.AsStringSlice(v); err != nil {
				return nil, fmt.Errorf(`failed to parse "ryftx-opts" option: %s`, err)
			} else {
				t.SetOptions("", "ryftx", "", vv)
			}
		}

		// `ryftpcre2` options
		if v, ok := opts_["ryftpcre2-opts"]; ok {
			if vv, err := utils.AsStringSlice(v); err != nil {
				return nil, fmt.Errorf(`failed to parse "ryftpcre2-opts" option: %s`, err)
			} else {
				t.SetOptions("", "ryftpcre2", "", vv)
			}
		}
	}

	// tweaks options
	opts, err := utils.AsStringMap(opts_["backend-tweaks"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "backend-tweaks" option: %s`, err)
	}

	// backend-tweaks.options
	if options_, ok := opts["options"]; ok {
		options, err := utils.AsStringMap(options_)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "backend-tweaks.options": %s`, err)
		}

		for k, v := range options {
			if vv, err := utils.AsStringSlice(v); err != nil {
				return nil, fmt.Errorf(`bad "backend-tweaks.options" value for key "%s": %s`, k, err)
			} else {
				t.Options[k] = vv
			}
		}
	}

	// backend-tweaks.router
	if router_, ok := opts["router"]; ok {
		router, err := utils.AsStringMap(router_)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse "backend-tweaks.router": %s`, err)
		}

		for k, v := range router {
			if tool, err := utils.AsString(v); err != nil {
				return nil, fmt.Errorf(`bad "backend-tweaks.router" value for key "%s": %s`, k, err)
			} else {
				// separator: space or any of ",;:"
				sep := func(r rune) bool {
					return unicode.IsSpace(r) ||
						strings.ContainsRune(",;:", r)
				}

				for _, kk := range strings.FieldsFunc(k, sep) {
					if mode := strings.TrimSpace(kk); len(mode) != 0 {
						t.Router[mode] = tool
					}
				}
			}
		}
	}

	return t, nil // OK
}

// GetOptions gets the custom backend options
func (t *Tweaks) GetOptions(mode, backend, primitive string) []string {
	// {mode.backend.primitive} combinations
	// in order of priority
	try := [][]string{
		[]string{mode, backend, primitive},
		[]string{mode, primitive},
		[]string{mode, backend},
		[]string{mode},
		[]string{backend, primitive},
		[]string{primitive},
		[]string{backend},
	}

	// check each combination
	for _, k := range try {
		key := strings.Join(k, ".")
		key = strings.TrimPrefix(key, ".")
		key = strings.TrimSuffix(key, ".")
		if len(key) == 0 {
			key = "default"
		}

		if v, ok := t.Options[key]; ok {
			return v
		}
	}

	return nil // not found
}

// SetOptions sets the custom backend options
func (t *Tweaks) SetOptions(mode, backend, primitive string, opts []string) {
	keys := []string{}
	if mode != "" {
		keys = append(keys, mode)
	}
	if backend != "" {
		keys = append(keys, backend)
	}
	if primitive != "" {
		keys = append(keys, primitive)
	}

	key := strings.Join(keys, ".")
	if opts != nil {
		t.Options[key] = opts
	} else {
		delete(t.Options, key)
	}
}

// GetBackendTool from routing table
func (t *Tweaks) GetBackendTool(primitive string) string {
	if tool, ok := t.Router[primitive]; ok {
		return tool
	}

	if tool, ok := t.Router["default"]; ok {
		return tool
	}

	return "" // not found
}
