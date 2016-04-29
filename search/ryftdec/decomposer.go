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

func Decompose(originalQuery string) (*Node, error) {
	rootNode := Node{SubNodes: make([]*Node, 0)}
	originalQuery = formatQuery(originalQuery)

	_, err := parse(&rootNode, originalQuery)
	if err != nil {
		return nil, err
	}
	node := normalizeTree(rootNode.SubNodes[0])

	return node, nil // Return first node with value
}

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
		return nil, buildError("Invalid query: " + query)
	}

	tokens := tokenize(query)

	if !validateTokens(tokens) {
		return nil, buildError("Invalid query: " + query)
	}

	tokens = translateToPrefixNotation(tokens)
	currentNode = addToTree(currentNode, tokens)

	if !validateTree(currentNode) {
		return nil, buildError("Invalid query: " + query)
	}

	return currentNode, nil
}

func normalizeTree(node *Node) *Node {
	if node.hasSubnodes() && node.sameTypeSubnodes() && node.subnodesAreQueries() {
		subnodesType := node.SubNodes[0].Type
		node.Expression = node.SubNodes[0].Expression + " " + node.Expression + " " + node.SubNodes[1].Expression
		node.Type = subnodesType
		node.SubNodes = node.SubNodes[0:0]
	} else {
		for _, subNode := range node.SubNodes {
			normalizeTree(subNode)
		}
	}

	return node
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
		result = append(result, tokens[index:]...)
		result = reorderOperators(tokens[:index], result)
	} else {
		result = append(result, tokens...)
	}

	return result
}

func addToTree(currentNode *Node, tokens []string) *Node {
	for _, token := range tokens {
		if isOperator(token) {
			currentNode = addChildToNode(currentNode, token)
		} else {
			if notParsable(token) {
				addChildToNode(currentNode, token)
			} else {
				_, _ = parse(currentNode, token)
			}
		}
	}
	return currentNode
}

func notParsable(expression string) bool {
	twoBrackets := (strings.Count(expression, "(") == 1) && (strings.Count(expression, ")") == 1)
	dateExpression := strings.Contains(expression, "DATE(")
	timeExpression := strings.Contains(expression, "TIME(")
	noBrackets := (strings.Count(expression, "(") == 0) && (strings.Count(expression, ")") == 0)
	return noBrackets || (twoBrackets && dateExpression) || (twoBrackets && timeExpression)
}

func addChildToNode(currentNode *Node, token string) *Node {
	// TODO: use New method to build node for expression
	newNode := nodeForExpression(token)
	currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	return &newNode
}

func nodeForExpression(expression string) Node {
	var newNode Node
	if isOperator(expression) {
		newNode = Node{Expression: strings.Trim(expression, " "), Type: operatorConst(expression)}
	} else {
		newNode = Node{Expression: "(" + expression + ")", Type: queryConst(expression)}
	}
	return newNode
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
	return containsString(delimiters, " "+token+" ")
}

func (node *Node) sameTypeSubnodes() bool {
	return node.SubNodes[0].Type == node.SubNodes[1].Type
}

func (node *Node) subnodesAreQueries() bool {
	// TODO: handle OR and XOR here
	return (node.SubNodes[0].Type != QTYPE_AND) && (node.SubNodes[1].Type != QTYPE_AND)
}

func (node *Node) hasSubnodes() bool {
	return len(node.SubNodes) > 0
}

func (node Node) String() string {
	return fmt.Sprintf("Expression: '%s'", node.Expression)
}
