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
	"errors"
	"fmt"
	"strings"
)

var (
	delimiters     = []string{" AND ", " OR "}
	markers        = []string{" DATE(", " TIME("}
	maxDepth   int = 1
)

type QueryType int

const (
	QTYPE_SEARCH QueryType = iota
	QTYPE_DATE
	QTYPE_TIME
	QTYPE_NUMERIC
	QTYPE_AND
	QTYPE_OR
	QTYPE_XOR
)

// IsSearch checks if query type is a search
func (q QueryType) IsSearch() bool {
	switch q {
	case QTYPE_SEARCH, QTYPE_DATE,
		QTYPE_TIME, QTYPE_NUMERIC:
		return true
	}

	return false
}

type Node struct {
	Expression string
	Type       QueryType
	SubNodes   []*Node
}

func (node Node) String() string {
	return fmt.Sprintf("Expression: '%s'", node.Expression)
}

func Decompose(originalQuery string) (*Node, error) {
	rootNode := Node{SubNodes: make([]*Node, 0)}
	originalQuery = formatQuery(originalQuery)

	_, err := parse(&rootNode, originalQuery)
	if err != nil {
		return nil, err
	}

	return rootNode.SubNodes[0], nil // Return first node with value
}

// Add spaces around logic operators
func formatQuery(query string) string {
	for _, delimiter := range delimiters {
		delimiter = strings.Trim(delimiter, " ")
		query = strings.Replace(query, ")"+delimiter+"(", ") "+delimiter+" (", -1)
	}
	query = strings.Replace(query, "  ", " ", -1)
	return query
}

// Parse expression and build query tree
func parse(currentNode *Node, query string) (*Node, error) {
	if !validateQuery(query) {
		return nil, buildError("Can't parse expression, invalid format")
	}

	tokens := tokenize(query)
	tokens = translateToPrefixNotation(tokens)
	currentNode = addToTree(currentNode, tokens)

	return currentNode, nil
}

func tokenize(query string) []string {
	count := 0
	isBracket := func(r rune) bool {
		switch {
		case r == '(':
			count++
			if count == maxDepth {
				return true
			} else {
				return false
			}
		case r == ')':
			count--
			if count == maxDepth-1 {
				return true
			} else {
				return false
			}
		default:
			return false
		}
	}
	return strings.FieldsFunc(query, isBracket)
}

func addToTree(currentNode *Node, tokens []string) *Node {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if isDecomposable(token) {
			parse(currentNode, token)
		} else {
			switch {
			case isOperator(token):
				currentNode = addChildToNode(currentNode, token)
			default:
				addChildToNode(currentNode, token)
			}
		}
	}
	return currentNode
}

// Decompose query only when it includes DATE/TIME operators and has logic operators AND/OR
func isDecomposable(originalQuery string) bool {
	return includesMultipleSearchTypes(originalQuery) && includesAnyToken(originalQuery, delimiters)
}

func includesMultipleSearchTypes(originalQuery string) bool {
	for _, marker := range markers {
		if containsMultipleTypes(originalQuery, marker) {
			return true
		}
	}
	return false
}

func formatSubQuery(query string) string {
	// Add brackets if query is not surrounded by them
	// e.g RAW_TEXT CONTAINS "100"
	if []rune(query)[0] != '(' {
		return "(" + query + ")"
	}
	return query
}

func containsMultipleExpressions(query string) bool {
	return includesAnyToken(query, delimiters)
}

// Check if expression has multiple kinds of expressions, e.g. (TEXT AND DATE) or maybe (DATE AND TIME)
func containsMultipleTypes(query string, marker string) bool {
	delimitersCount := 0
	for _, delimiter := range delimiters {
		count := strings.Count(query, delimiter)
		delimitersCount = delimitersCount + count
	}

	markersCount := strings.Count(query, marker)

	return (delimitersCount == markersCount) || (delimitersCount > 1 && markersCount > 0 && markersCount < delimitersCount)
}

func includesAnyToken(query string, tokens []string) bool {
	for _, marker := range tokens {
		if strings.Contains(query, marker) {
			return true
		}
	}
	return false
}

func translateToPrefixNotation(tokens []string) []string {
	for i := 0; i < len(tokens)-1; i++ {
		if isOperator(tokens[i]) {
			tokens[i-1], tokens[i] = tokens[i], tokens[i-1]
		}
	}
	return tokens
}

func addChildToNode(currentNode *Node, token string) *Node {
	var newNode Node
	switch {
	case isOperator(token):
		newNode = Node{Expression: strings.Trim(token, " "), Type: operatorConst(token)}
	default:
		newNode = Node{Expression: formatSubQuery(token), Type: queryConst(token)}
	}
	currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	return &newNode
}

// Map string operator value to constant
func operatorConst(token string) QueryType {
	token = strings.Trim(token, " ")
	switch token {
	case "AND":
		return QTYPE_AND
	case "OR":
		return QTYPE_OR
	default:
		return QTYPE_XOR
	}
}

func queryConst(query string) QueryType {
	switch {
	case strings.Contains(query, "DATE("):
		return QTYPE_DATE
	case strings.Contains(query, "TIME("):
		return QTYPE_TIME
		//case strings.Contains(query, "????"):
		//return QTYPE_NUMERIC
	}

	return QTYPE_SEARCH
}

func isOperator(token string) bool {
	return containsString(delimiters, token)
}

func buildError(message string) error {
	return errors.New(message)
}
