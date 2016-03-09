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

package utils

import (
	"fmt"
	"strconv"
	"time"
)

// convert custom value to string.
func AsString(opt interface{}) (string, error) {
	switch v := opt.(type) {
	// TODO: other types to string?
	case string:
		return v, nil
	case nil:
		return "", nil
	}

	return "", fmt.Errorf("%v is not a string", opt)
	// return fmt.Sprintf("%s", opt), nil
}

// convert custom value to time duration.
func AsDuration(opt interface{}) (time.Duration, error) {
	switch v := opt.(type) {
	// TODO: other types to duration?
	case string:
		return time.ParseDuration(v)
	case time.Duration:
		return v, nil
	case nil:
		return time.Duration(0), nil
	}

	return time.Duration(0), fmt.Errorf("%v is not a time duration", opt)
}

// convert custom value to uint64.
func AsUint64(opt interface{}) (uint64, error) {
	switch v := opt.(type) {
	// TODO: other types to uint64?
	case uint:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int64:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	case nil:
		return 0, nil
	}

	return 0, fmt.Errorf("%v is not an uint64", opt)
}

// convert custom value to bool.
func AsBool(opt interface{}) (bool, error) {
	switch v := opt.(type) {
	// TODO: other types to bool?
	case bool:
		return v, nil
	case nil:
		return false, nil
	}

	return false, fmt.Errorf("%v is not a bool", opt)
}
