// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binding

import (
	"fmt"
	"log"
	"strconv"

	"github.com/getryft/ryft-rest-api/ryft-server/formats"
	"github.com/gin-gonic/gin"
)

type Search struct {
	Query       string   `form:"query" json:"query" binding:"required"`             // For example: ( RAW_TEXT CONTAINS "night" )
	Files       []string `form:"files" json:"files" binding:"required"`             // Splitted OS-specific ListSeparator: "/a/b/c:/usr/bin/file" -> "/a/b/c", "/usr/bin/file"
	Surrounding uint16   `form:"surrounding" json:"surrounding" binding:"required"` // Specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text
	Fuzziness   uint8    `form:"fuzziness" json:"fuzziness"`                        // Is the fuzziness of the search. Measured as the maximum Hamming distance.
	Format      string   `form:"format" json:"format"`                              // Source format parser
	Out         string
}

const (
	queryTag       = "query"
	filesTag       = "files"
	surroundingTag = "surrounding"
	fuzzinessTag   = "fuzziness"
	formatTag      = "format"
)

func NewSearch(c *gin.Context) (*Search, error) {
	s := new(Search)
	url := c.Request.URL.Query()

	if c.Request.Header.Get("Content-Type") == "application/msgpk" ||
		c.Request.Header.Get("Content-Type") == "application/x-msgpk" {
		s.Out = "msgpk"
	} else {
		s.Out = "json"

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

	return s, nil
}
