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
	"strings"

	"github.com/getryft/ryft-server/search"
)

var (
	delimiters     = []string{" AND ", " OR "}
	markers        = []string{" DATE(", " TIME(", " NUMBER(", " FHS(", " FEDS(", " CURRENCY(", "REGEX(", "IPV4("}
	maxDepth   int = 1
)

// Options contains search options
type Options struct {
	Mode                  string         // Search mode: fhs, feds, date, time, etc.
	Width                 uint           // Surrounding width
	Dist                  uint           // Fuzziness distance
	Cs                    bool           // Case sensitivity flag
	BooleansPerExpression map[string]int // Number of permitted booleans per expression type
}

// convert search config to Options
func configToOpts(config *search.Config) Options {
	return Options{
		Mode:  config.Mode,
		Dist:  config.Fuzziness,
		Width: config.Surrounding,
		Cs:    config.CaseSensitive,
	}
}

// Decompose search expression into expression tree
func Decompose(originalQuery string, baseOpts Options) (node *Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Decomposer: %v", r)
			}
		}
	}()

	rootNode := Node{SubNodes: make([]*Node, 0)}
	originalQuery = formatQuery(originalQuery)

	parse(&rootNode, originalQuery, baseOpts)

	if err != nil {
		return nil, err
	}

	node = rootNode.SubNodes[0]
	normalizeTree(node, baseOpts.BooleansPerExpression)

	return
}

func formatQuery(query string) string {
	query = strings.TrimSpace(query)
	for _, delimiter := range delimiters {
		delimiter = strings.Trim(delimiter, " ")
		query = strings.Replace(query, ")"+delimiter+"(", ") "+delimiter+" (", -1)
	}
	query = strings.Replace(query, "  ", " ", -1)
	return query
}

// Parse expression and build query tree
func parse(currentNode *Node, query string, opts Options) {
	if !validateQuery(query) {
		panic(fmt.Errorf("Invalid query: %q", query))
	}

	tokens := tokenize(query)

	if !validateTokens(tokens) {
		panic(fmt.Errorf("Invalid query: %q (bad tokens)", query))
	}

	tokens = translateToPrefixNotation(tokens)
	currentNode = addToTree(currentNode, tokens, opts)

	if !validateTree(currentNode) {
		panic(fmt.Errorf("Invalid query: %q (bad tree)", query))
	}
}

func tokenize(query string) []string {
	count := 0
	quotesCount := 0

	isBracket := func(r rune) bool {
		switch {
		case r == '(':
			if quotesCount == 0 {
				count++
			}
			if count == maxDepth {
				return true
			} else {
				return false
			}
		case r == ')':
			if quotesCount == 0 {
				count--
			}
			if count == maxDepth-1 {
				return true
			} else {
				return false
			}
		default:
			return false
		}
	}

	tokens := strings.FieldsFunc(query, isBracket)
	for i, token := range tokens {
		tokens[i] = strings.Trim(token, " ")
	}
	return tokens
}

func translateToPrefixNotation(tokens []string) []string {
	if containsString(tokens, "OR") && containsString(tokens, "AND") {
		tokens = reorderOperators(tokens, make([]string, 0))
	}

	for i := 1; i < len(tokens)-1; i++ {
		if isOperator(tokens[i]) {
			tokens[i-1], tokens[i] = tokens[i], tokens[i-1]
		}
	}
	return tokens
}

func reorderOperators(tokens []string, result []string) []string {
	index := indexOfToken(tokens, "OR")
	if index > 0 {
		result = append(result, tokens[index])
		result = append(result, tokens[:index]...)
		result = append(result, tokens[index+1:]...)
	} else {
		result = append(result, tokens...)
	}

	return result
}

func addToTree(currentNode *Node, tokens []string, opts Options) *Node {
	for _, token := range tokens {
		if notParsable(token) {
			currentNode = addChildToNode(currentNode, token, opts)
		} else {
			parse(currentNode, token, opts)
		}
	}
	return currentNode
}

func addChildToNode(currentNode *Node, expression string, opts Options) *Node {
	var node *Node
	if len(currentNode.SubNodes) == 2 {
		node = NewNode(expression, currentNode.Parent, opts)
		currentNode.Parent.SubNodes = append(currentNode.Parent.SubNodes, node)
	} else {
		node = NewNode(expression, currentNode, opts)
		currentNode.SubNodes = append(currentNode.SubNodes, node)
	}

	if isOperator(expression) {
		return node
	} else {
		return currentNode
	}
}

func notParsable(expression string) bool {
	expression = removeQuotedText(expression)
	twoBrackets := (strings.Count(expression, "(") == 1) && (strings.Count(expression, ")") == 1)
	noBrackets := (strings.Count(expression, "(") == 0) && (strings.Count(expression, ")") == 0)

	return noBrackets || (twoBrackets && isSearchQuery(expression))
}

func isSearchQuery(expression string) bool {
	for _, v := range markers {
		if strings.Contains(expression, v) {
			return true
		}
	}
	return false
}

func isOperator(token string) bool {
	return containsString(delimiters, " "+token+" ")
}
