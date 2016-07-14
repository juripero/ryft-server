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

package ryftdec

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type QueryType int

const (
	QTYPE_SEARCH QueryType = iota
	QTYPE_DATE
	QTYPE_TIME
	QTYPE_NUMERIC
	QTYPE_CURRENCY
	QTYPE_REGEX

	QTYPE_AND
	QTYPE_OR
	QTYPE_XOR
)

type Options struct {
	Expression    string
	Values        map[string]interface{}
	DefaultValues *GlobalOptions
}

func (o *Options) Cs() (value bool) {
	val := o.Values["cs"]
	cs, ok := val.(bool)

	if ok {
		return cs
	} else {
		return o.DefaultValues.Cs
	}
}

func (o *Options) Dist() (value uint) {
	val := o.Values["dist"]
	dist, ok := val.(uint)

	if ok {
		return dist
	} else {
		return o.DefaultValues.Dist
	}
}

func (o *Options) Width() (value uint) {
	val := o.Values["width"]
	width, ok := val.(uint)

	if ok {
		return width
	} else {
		return o.DefaultValues.Width
	}
}

func NewOptions(expression string, globalOpts GlobalOptions) Options {
	expr, optionValues := parseOptions(expression)

	if expressionType(expr).IsSearch() {
		expr = fmt.Sprint("(", expr, ")")
	}

	return Options{Expression: expr, Values: optionValues, DefaultValues: &globalOpts}
}

func parseOptions(expression string) (cleanExpression string, values map[string]interface{}) {
	values = make(map[string]interface{})
	cleanExpression = expression

	regex := regexp.MustCompile(`\(?(.+) (FHS|FEDS)\((.+?),?\s?([\s\w]+)?,?\s?(\d*)?,?\s?(\d*)?\)`)
	matches := regex.FindAllStringSubmatch(expression, -1)

	if len(matches) > 0 {
		match := matches[0]

		if match[4] != "" {
			cs, err := strconv.ParseBool(strings.TrimSpace(match[4]))
			if err != nil {
				panic(err)
			}
			values["cs"] = cs
		}

		if match[5] != "" {
			dist64, err := strconv.ParseInt(strings.TrimSpace(match[5]), 10, 0)
			if err != nil {
				panic(err)
			}
			values["dist"] = int(dist64)
		}

		if match[6] != "" {
			width64, err := strconv.ParseInt(strings.TrimSpace(match[6]), 10, 0)
			if err != nil {
				panic(err)
			}
			values["width"] = int(width64)
		}

		cleanExpression = fmt.Sprint(match[1], " ", match[3])
	}

	return
}

type Node struct {
	//Expression string
	Type     QueryType
	Parent   *Node
	SubNodes []*Node
	Options
}

func (node *Node) New(expression string, parent *Node, globalOpts GlobalOptions) *Node {
	node.Options = NewOptions(expression, globalOpts)
	node.Type = expressionType(expression)
	node.Parent = parent
	return node
}

func (node *Node) sameTypeSubnodes() bool {
	return node.SubNodes[0].Type == node.SubNodes[1].Type
}

func (node *Node) subnodesAreQueries() bool {
	return node.SubNodes[0].Type.IsSearch() && node.SubNodes[1].Type.IsSearch()
}

func (node *Node) hasSubnodes() bool {
	return len(node.SubNodes) == 2
}

func (node Node) String() string {
	return fmt.Sprintf("Expression: '%s'", node.Expression)
}

func (node *Node) isSearch() bool {
	return node.Type.IsSearch()
}

func (node *Node) isOperator() bool {
	return !node.Type.IsSearch()
}

// Map string operator value to constant
func expressionType(expression string) QueryType {
	expression = strings.Trim(expression, " ")
	switch {
	case expression == "AND":
		return QTYPE_AND
	case expression == "OR":
		return QTYPE_OR
	case expression == "XOR":
		return QTYPE_XOR
	case strings.Contains(expression, "DATE("):
		return QTYPE_DATE
	case strings.Contains(expression, "TIME("):
		return QTYPE_TIME
	case strings.Contains(expression, "NUMBER("):
		return QTYPE_NUMERIC
	case strings.Contains(expression, "CURRENCY("):
		return QTYPE_CURRENCY
	case strings.Contains(expression, "REGEX("):
		return QTYPE_REGEX
	default:
		return QTYPE_SEARCH
	}
}

// IsSearch checks if query type is a search
func (q QueryType) IsSearch() bool {
	switch q {
	case QTYPE_SEARCH, QTYPE_DATE,
		QTYPE_TIME, QTYPE_NUMERIC, QTYPE_CURRENCY, QTYPE_REGEX:
		return true
	}

	return false
}
