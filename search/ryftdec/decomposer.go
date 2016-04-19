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
	delimiters = []string{" AND ", " OR "}
	markers    = []string{" DATE(", " TIME("}
)

type SubQuery struct {
	query    string
	operator string
}

func (subquery SubQuery) String() string {
	return fmt.Sprintf("Query: %s Operator: %s}", subquery.query, subquery.operator)
}

func decompose(originalQuery string) []SubQuery {
	queries := make([]SubQuery, 0)
	// load tokens one by one
	// each token is logic operator or (expression) that should not be decomposed any more
	operators := parse(make([]string, 0), originalQuery)

	// build SubQuery instances attaching next logic operator to current query
	for i := 0; i <= len(operators)-1; i = i + 2 {
		var operator string
		if i < len(operators)-1 {
			operator = operators[i+1]
		}
		queries = append(queries, SubQuery{query: operators[i], operator: operator})
	}

	return queries
}

// Decompose query only when it includes DATE/TIME operators and has logic operators AND/OR
func isDecomposable(originalQuery string) bool {
	return includesAnyToken(originalQuery, delimiters) && includesAnyToken(originalQuery, markers)
}

func splitQuery(queries []SubQuery, originalQuery string) []SubQuery {
	return queries
}

func includesAnyToken(query string, tokens []string) bool {
	for _, marker := range tokens {
		if strings.Contains(query, marker) {
			return true
		}
	}
	return false
}

func parse(queries []string, str string) []string {
	depth := 1
	count := 0
	isBracket := func(r rune) bool {
		switch {
		case r == '(':
			count++
			if count == depth {
				return true
			} else {
				return false
			}
		case r == ')':
			count--
			if count == depth-1 {
				return true
			} else {
				return false
			}
		default:
			return false
		}
	}

	args := strings.FieldsFunc(str, isBracket)
	for _, t := range args {
		if isDecomposable(t) {
			queries = parse(queries, t)
		} else {
			queries = append(queries, t)
		}
	}
	return queries
}
