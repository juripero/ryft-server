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
	QTYPE_AND
	QTYPE_OR
	QTYPE_XOR
)

type Options struct {
	Expression string
	Cs         bool
	Dist       int
	Width      int
}

func NewOptions(expression string) Options {
	expr, cs, dist, width := parseOptions(expression)

	if expressionType(expr).IsSearch() {
		expr = fmt.Sprint("(", expr, ")")
	}

	return Options{Expression: expr, Cs: cs, Dist: dist, Width: width}
}

func parseOptions(expression string) (string, bool, int, int) {
	var (
		cs    bool
		dist  int
		width int
		err   error
	)

	regex := regexp.MustCompile(`(.+) (FHS|FEDS)\((([\"\']{1}.+[\"\']{1}),?(.+)?)\)`)
	matches := regex.FindAllStringSubmatch(expression, -1)

	if len(matches) > 0 {
		// Capture search query
		args := strings.Split(matches[0][3], ",")

		if len(args) > 1 {
			cs, err = strconv.ParseBool(strings.TrimSpace(args[1]))
			if err != nil {
				panic(err)
			}
		}

		if len(args) > 2 {
			dist64, err := strconv.ParseInt(strings.TrimSpace(args[2]), 10, 0)
			if err != nil {
				panic(err)
			}
			dist = int(dist64)
		}

		if len(args) > 3 {
			width64, err := strconv.ParseInt(strings.TrimSpace(args[3]), 10, 0)
			if err != nil {
				panic(err)
			}
			width = int(width64)
		}

		expression = fmt.Sprint(matches[0][1], " ", matches[0][4])
	}

	return expression, cs, dist, width
}

type Node struct {
	//Expression string
	Type     QueryType
	Parent   *Node
	SubNodes []*Node
	Options
}

func (node *Node) New(expression string, parent *Node) *Node {
	node.Options = NewOptions(expression)
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
	default:
		return QTYPE_SEARCH
	}
}

// IsSearch checks if query type is a search
func (q QueryType) IsSearch() bool {
	switch q {
	case QTYPE_SEARCH, QTYPE_DATE,
		QTYPE_TIME, QTYPE_NUMERIC, QTYPE_CURRENCY:
		return true
	}

	return false
}
