/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

// AsString converts custom value to string.
func AsString(opt interface{}) (string, error) {
	switch v := opt.(type) {
	// TODO: other types to string?
	case nil:
		return "", nil
	case string:
		return v, nil
	}

	return "", fmt.Errorf("%v is not a string", opt)
	// return fmt.Sprintf("%s", opt), nil
}

// AsDuration converts custom value to time duration.
func AsDuration(opt interface{}) (time.Duration, error) {
	switch v := opt.(type) {
	// TODO: other types to duration?
	case nil:
		return time.Duration(0), nil
	case string:
		return time.ParseDuration(v)
	case time.Duration:
		return v, nil
	}

	return time.Duration(0), fmt.Errorf("%v is not a time duration", opt)
}

// AsInt64 converts custom value to int64.
func AsInt64(opt interface{}) (int64, error) {
	switch v := opt.(type) {
	// TODO: other types to uint64?
	case nil:
		return 0, nil
	case uint:
		return int64(v), nil
	case int:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case int64:
		return v, nil
	case uint32:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	}

	return 0, fmt.Errorf("%v is not an int64", opt)
}

// AsUint64 converts custom value to uint64.
func AsUint64(opt interface{}) (uint64, error) {
	switch v := opt.(type) {
	// TODO: other types to uint64?
	case nil:
		return 0, nil
	case uint:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int64:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case float32:
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	}

	return 0, fmt.Errorf("%v is not an uint64", opt)
}

// AsBool converts custom value to bool.
func AsBool(opt interface{}) (bool, error) {
	switch v := opt.(type) {
	// TODO: other types to bool?
	case nil:
		return false, nil
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	}

	return false, fmt.Errorf("%v is not a bool", opt)
}

// AsFloat64 converts custom value to float64.
func AsFloat64(opt interface{}) (float64, error) {
	switch v := opt.(type) {
	// TODO: other types to float64?
	case nil:
		return 0.0, nil
	case uint:
		return float64(v), nil
	case int:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case float64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	}

	return 0, fmt.Errorf("%v is not a float64", opt)
}

// AsStringSlice get value as string slice
func AsStringSlice(v interface{}) ([]string, error) {
	switch vv := v.(type) {
	case nil:
		return nil, nil // empty

	case string:
		return []string{vv}, nil // OK, single element

	case []string:
		return vv, nil // OK

	case []interface{}:
		res := make([]string, 0, len(vv))
		for _, s := range vv {
			if ss, err := AsString(s); err != nil {
				return nil, fmt.Errorf("not a string: %s", err)
			} else {
				res = append(res, ss)
			}
		}
		return res, nil // OK
	}

	return nil, fmt.Errorf("not a []string (but %T)", v)
}

// AsStringMap get value as map[string]interface{}
func AsStringMap(v interface{}) (map[string]interface{}, error) {
	switch vv := v.(type) {
	case nil:
		return nil, nil // empty

	case map[string]interface{}:
		return vv, nil // OK

	case map[interface{}]interface{}:
		res := make(map[string]interface{}, len(vv))
		for k, v := range vv {
			if kk, err := AsString(k); err != nil {
				return nil, fmt.Errorf("bad key: %s", err)
			} else {
				res[kk] = v
			}
		}
		return res, nil // OK
	}

	return nil, fmt.Errorf("not a map[string]interface{} (but %T)", v)
}
