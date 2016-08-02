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

// parses options from search expression
func parseExpression(expression string, baseOpts Options) (cleanedExpression string, opts Options) {
	opts = baseOpts // just a copy by default
	cleanedExpression = expression

	regex := regexp.MustCompile(`\(?(.+) (FHS|FEDS)\(([\"\'].+?[\"\']),?(.+?)?\)\)?`)
	matches := regex.FindAllStringSubmatch(expression, -1)

	if len(matches) > 0 {
		match := matches[0]

		op := strings.TrimSpace(match[1])
		mode := strings.TrimSpace(match[2])
		expr := strings.TrimSpace(match[3])
		argsString := strings.TrimSpace(match[4])
		args := parseArgs(argsString)

		opts.Mode = strings.ToLower(mode) // FHS or FEDS

		for name, value := range args {
			opts = SetOption(opts, name, value)
		}

		// remove all embedded options from search expression
		cleanedExpression = fmt.Sprintf("%s %s", op, expr)
	}

	return
}

func parseArgs(argsString string) map[string]string {
	args := make(map[string]string)
	regex := regexp.MustCompile(`(:?(\w+)=([\w\d]+))`)
	matches := regex.FindAllStringSubmatch(argsString, -1)

	for _, match := range matches {
		key := match[2]
		value := match[3]
		args[key] = value
	}
	return args
}

func SetOption(o Options, name, value string) Options {
	switch name {
	case "CS":
		v, err := strconv.ParseBool(value)
		if err != nil {
			panic(err)
		}
		o.Cs = v
	case "DIST":
		v, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			panic(err)
		}
		o.Dist = uint(v)
	case "WIDTH":
		v, err := strconv.ParseInt(strings.TrimSpace(value), 10, 0)
		if err != nil {
			panic(err)
		}
		o.Width = uint(v)
	}

	return o
}

// Search tree node
type Node struct {
	Expression string
	Type       QueryType
	Parent     *Node
	SubNodes   []*Node
	Options    Options
}

// create new tree node
func NewNode(expression string, parent *Node, baseOpts Options) *Node {
	node := new(Node)
	node.Expression, node.Options = parseExpression(expression, baseOpts)
	node.Type = expressionType(node.Expression)
	if node.Type.IsSearch() {
		node.Expression = fmt.Sprintf("(%s)", node.Expression)
	}
	node.Parent = parent
	return node
}

func (node *Node) isNormalizable() bool {
	return node.hasSubnodes() &&
		node.sameTypeSubnodes() &&
		node.subnodesAreQueries() &&
		node.boolLimitIsNotReached()
}

func (node *Node) boolLimitIsNotReached() bool {
	boolCount := 0
	for _, subNode := range node.SubNodes {
		boolCount += countBoolOperators(subNode)
	}
	return boolCount == 0
}

func (node *Node) sameTypeSubnodes() bool {
	return node.hasSubnodes() &&
		(node.SubNodes[0].Type == node.SubNodes[1].Type) &&
		node.SubNodes[0].optionsEqual(node.SubNodes[1])
}

func (node *Node) subnodesAreQueries() bool {
	return node.SubNodes[0].Type.IsSearch() && node.SubNodes[1].Type.IsSearch()
}

func (node *Node) hasSubnodes() bool {
	return len(node.SubNodes) == 2
}

func (node Node) String() string {
	return fmt.Sprintf("Expression: %q", node.Expression)
}

func (node *Node) isSearch() bool {
	return node.Type.IsSearch()
}

func (node *Node) isOperator() bool {
	return !node.Type.IsSearch()
}

func (node *Node) optionsEqual(cmpNode *Node) bool {
	a := node.Options
	b := cmpNode.Options
	return (a.Mode == b.Mode) &&
		(a.Dist == b.Dist) &&
		(a.Width == b.Width) &&
		(a.Cs == b.Cs)
}

// Map string operator value to constant
func expressionType(expression string) QueryType {
	expression = removeQuotedText(strings.Trim(expression, " "))
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
		QTYPE_TIME, QTYPE_NUMERIC,
		QTYPE_CURRENCY, QTYPE_REGEX:
		return true
	}

	return false
}
