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

package binding

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/getryft/ryft-server/formats"
	"github.com/getryft/ryft-server/records"
	"github.com/gin-gonic/gin"
)

type Search struct {
	Query           string                                         // For example: ( RAW_TEXT CONTAINS "night" )
	Files           []string                                       // Source files
	Surrounding     uint16                                         // Specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text
	Fuzziness       uint8                                          // Is the fuzziness of the search. Measured as the maximum Hamming distance.
	Format          string                                         // Source format parser name
	FormatConvertor func(r records.IdxRecord) (interface{}, error) // Source format parser (calculating from Format)
	CaseSensitive   bool                                           // Case sensitive flag
	out             string                                         // Output format in header (msgpack or json)
	State           int                                            // Output State; for output array wrapper
}

const (
	StateBegin = iota
	StateBody  = iota
	StateEnd   = iota
)

const (
	outJson  = "json"
	outMsgpk = "msgpk"
)

var (
	yesValues = []string{"1", "y", "yes", "t", "true"}
	noValues  = []string{"0", "n", "no", "f", "false"}
)

func isAnyStr(s string, ss []string) bool {
	ls := strings.ToLower(s)
	for _, t := range ss {
		if ls == t {
			return true
		}
	}
	return false
}

func isYes(s string) bool {
	return isAnyStr(s, yesValues)
}

func isNo(s string) bool {
	return isAnyStr(s, noValues)
}

const (
	queryTag         = "query"
	filesTag         = "files"
	surroundingTag   = "surrounding"
	fuzzinessTag     = "fuzziness"
	formatTag        = "format"
	caseSensitiveTag = "cs"
)

func NewSearch(c *gin.Context) (*Search, error) {
	s := new(Search)
	s.State = StateBegin

	url := c.Request.URL.Query()

	if c.Request.Header.Get("Content-Type") == "application/msgpack" ||
		c.Request.Header.Get("Content-Type") == "application/x-msgpack" {
		s.out = outMsgpk
	} else {
		s.out = outJson
	}

	query, hasQuery := url[queryTag]
	if !hasQuery || len(query) != 1 {
		return nil, fmt.Errorf("Query can not be empty. Query should be single.")
	}
	s.Query = query[0]

	files, hasFiles := url[filesTag]
	if !hasFiles || len(files) == 0 {
		return nil, fmt.Errorf("At least one file should be specified")
	}
	s.Files = files

	surrounding, hasSurrounding := url[surroundingTag]
	if !hasSurrounding {
		s.Surrounding = 0
	} else {
		srValue, srValueErr := strconv.ParseInt(surrounding[0], 0, 64)

		if srValueErr != nil {
			return nil, fmt.Errorf("Could not parse Surrounding")
		} else {
			if srValue < 0 {
				return nil, fmt.Errorf("Surrounding should not be less than 0")
			}

			if srValue > 32767 {
				return nil, fmt.Errorf("Surrounding should not be more than 32767")
			}

		}
		s.Surrounding = uint16(srValue)
	}

	fuzziness, hasFuzziness := url[fuzzinessTag]
	if !hasFuzziness {
		s.Fuzziness = 0
	} else {
		fzValue, fzValueErr := strconv.ParseInt(fuzziness[0], 0, 64)
		if fzValueErr != nil {
			return nil, fmt.Errorf("Could not parse Fuzziness")
		} else {
			if fzValue < 0 {
				return nil, fmt.Errorf("Fuzziness should not be less than 0")
			}

			if fzValue > 254 {
				log.Println("inside")
				return nil, fmt.Errorf("Fuzziness should not be more than 254")
			}
		}
		s.Fuzziness = uint8(fzValue)
	}

	format, hasFormat := url[formatTag]
	if !hasFormat {
		s.Format = formats.Default()
	} else {
		if !formats.Available(format[0]) {
			return nil, fmt.Errorf("Parsing format has not supported")
		}
		s.Format = format[0]
	}

	s.FormatConvertor = formats.Formats()[s.Format]

	cs, hasCs := url[caseSensitiveTag]
	if hasCs {
		if isNo(cs[0]) {
			s.CaseSensitive = false
		} else if isYes(cs[0]) {
			s.CaseSensitive = true
		} else {
			return nil, fmt.Errorf(`Supported cs (Case Sensitivity) values: "1", "y", "yes", "t", "true", "0", "n", "no", "f", "false"`)
		}

	} else {
		s.CaseSensitive = true
	}

	return s, nil
}

func (s *Search) IsOutJson() bool {
	return s.out == outJson
}

func (s *Search) IsOutMsgpk() bool {
	return s.out == outMsgpk
}
