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

	"github.com/getryft/ryft-server/search/utils"
)

// Tweaks: absolute path flags
type Tweaks struct {
	// absolute path: [backend/tool] => flag
	UseAbsPath map[string]bool
}

// ParseTweaks parses tweaks from engine options
func ParseTweaks(opts_ map[string]interface{}) (*Tweaks, error) {
	t := new(Tweaks)
	t.UseAbsPath = make(map[string]bool)

	// tweaks options
	opts, err := utils.AsStringMap(opts_["backend-tweaks"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "backend-tweaks" option: %s`, err)
	}

	// backend-tweaks.abs-path
	if absPath_, ok := opts["abs-path"]; ok {
		var asMap map[string]interface{}
		var asSlice []string

		// multi-type option
		switch vv := absPath_.(type) {
		case nil:
			break // no configuration

		case map[string]interface{}:
			asMap = vv // map

		case map[interface{}]interface{}:
			if m, err := utils.AsStringMap(vv); err != nil {
				return nil, fmt.Errorf(`failed to parse "backend-tweaks.abs-path": %s`, err)
			} else {
				asMap = m
			}

		case []string:
			asSlice = vv // slice

		case []interface{}:
			if a, err := utils.AsStringSlice(vv); err != nil {
				return nil, fmt.Errorf(`failed to parse "backend-tweaks.abs-path": %s`, err)
			} else {
				asSlice = a
			}

		case string:
			asSlice = []string{vv} // slice of one element

		case bool:
			asMap = map[string]interface{}{
				"default": vv,
			}

		default:
			return nil, fmt.Errorf(`unknown "backend-tweaks.abs-path" option type: %T`, absPath_)
		}

		if asMap != nil {
			for k, v := range asMap {
				if flag, err := utils.AsBool(v); err != nil {
					return nil, fmt.Errorf(`bad "backend-tweaks.abs-path" value for key "%s": %s`, k, err)
				} else {
					t.UseAbsPath[k] = flag
				}
			}
		} else if asSlice != nil {
			for _, k := range asSlice {
				t.UseAbsPath[k] = true
			}
		}
	}

	return t, nil // OK
}
