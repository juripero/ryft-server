// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binding

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Search struct {
	Query       string   `form:"query" json:"query" binding:"required"`             // For example: ( RAW_TEXT CONTAINS "night" )
	Files       []string `form:"files" json:"files" binding:"required"`             // Splitted OS-specific ListSeparator: "/a/b/c:/usr/bin/file" -> "/a/b/c", "/usr/bin/file"
	Surrounding uint16   `form:"surrounding" json:"surrounding" binding:"required"` // Specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text
	Fuzziness   uint8    `form:"fuzziness" json:"fuzziness"`                        // Is the fuzziness of the search. Measured as the maximum Hamming distance.
}

const (
	queryTag       = "query"
	filesTag       = "files"
	surroundingTag = "surrounding"
	fuzzinessTag   = "fuzziness"
)

func NewSearch(c *gin.Context) (*Search, error) {
	s := new(Search)

	url := c.Request.URL.Query()

	query, hasQuery := url[queryTag]
	files, hasFiles := url[filesTag]
	surrounding, hasSurrounding := url[surroundingTag]
	fuzziness, hasFuzziness := url[fuzzinessTag]

	if !hasQuery || len(query) <= 0 {
		return nil, fmt.Errorf("Query can not be empty")
	}

	if !hasFiles || len(files) <= 0 {
		return nil, fmt.Errorf("At least one file should be specified")
	}

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
	}

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
				return nil, fmt.Errorf("Fuzziness should not be more than 254 ")
			}
		}
	}
	c.Bind(s)
	return s, nil

}
