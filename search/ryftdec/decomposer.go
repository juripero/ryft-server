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

type Node struct {
	query    string
	operator string
	nodeType string
	subNodes []*Node
}

func (node Node) String() string {
	if node.nodeType == "operator" {
		return fmt.Sprintf("Op: '%s'", node.operator)
	} else {
		return fmt.Sprintf("Q: '%s'", node.query)
	}
}

func decompose(originalQuery string) *Node {
	rootNode := Node{nodeType: "root", subNodes: make([]*Node, 0)}
	originalQuery = formatQuery(originalQuery)
	parse(&rootNode, originalQuery)
	return &rootNode
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
func parse(currentNode *Node, str string) *Node {
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

	var token string
	tokens := strings.FieldsFunc(str, isBracket)
	tokens = translateToPrefixNotation(tokens)

	// Build tree from tokens
	for i := 0; i < len(tokens); i++ {
		token = tokens[i]
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
	return includesAnyToken(originalQuery, delimiters) && includesAnyToken(originalQuery, markers)
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
		newNode = Node{operator: token, nodeType: "operator"}
	default:
		newNode = Node{query: token, nodeType: "query"}
	}
	currentNode.subNodes = append(currentNode.subNodes, &newNode)
	return &newNode
}

func isOperator(token string) bool {
	return containsString(delimiters, token)
}
