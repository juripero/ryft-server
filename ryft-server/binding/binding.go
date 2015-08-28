// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binding

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type Search struct {
	Query       string   `form:"query" json:"query" binding:"required"`             // For example: ( RAW_TEXT CONTAINS "night" )
	Files       []string `form:"files" json:"files" binding:"required"`             // Splitted OS-specific ListSeparator: "/a/b/c:/usr/bin/file" -> "/a/b/c", "/usr/bin/file"
	Surrounding uint16   `form:"surrounding" json:"surrounding" binding:"required"` // Specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text
	Fuzziness   uint8    `form:"fuzziness" json:"fuzziness"`                        // Is the fuzziness of the search. Measured as the maximum Hamming distance.
}

func NewSearch(c *gin.Context) (*Search, error) {
	s := new(Search)
	if err := c.Bind(s); err != nil {
		log.Printf("SEARCH = %+v", s)
		if s.Query == "" {
			return nil, fmt.Errorf("Query can not be empty")
		}
		if len(s.Files) <= 0 {
			return nil, fmt.Errorf("At least one file should be specified")
		}
		if s.Surrounding <= 0 && s.Surrounding <= 32768 {
			return nil, fmt.Errorf("Surrounding should not be less than 0 and not more than 32768")
		}
		if s.Fuzziness <= 0 && s.Fuzziness <= 255 {
			return nil, fmt.Errorf("Fuzziness should not be less than 0 and not more than 255")
		}
	}

	return s, nil

}
