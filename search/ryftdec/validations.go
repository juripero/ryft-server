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
	"regexp"
	"strings"
)

func validateQuery(query string) bool {
	chars := []rune(query)

	if !validateQueryLength(chars) {
		return false
	}

	if !validateEmptyBrackets(query) {
		return false
	}

	if !validateBracketsBalance(chars) {
		return false
	}

	return true
}

func validateBracketsBalance(chars []rune) bool {
	count := 0
	for i := 0; i < len(chars); i++ {
		c := chars[i]

		if count < 0 {
			return false
		}

		switch {
		case c == '(':
			count++
		case c == ')':
			count--
		}
	}

	return count == 0
}

func validateQueryLength(chars []rune) bool {
	return len(chars) > 2
}

// Check if empty brackets are surrounded by quotes
func validateEmptyBrackets(query string) bool {
	bracketsPresent := strings.Contains(query, "()")
	quotedBrackets, err := regexp.MatchString(`[^".]+?\(\)[^".]*?`, query)
	if err != nil {
		return false
	}
	return (bracketsPresent && quotedBrackets) || !bracketsPresent
}

func validateTokens(tokens []string) bool {
	for _, token := range tokens {
		if notParsable(token) && !isOperator(token) && !validateToken(token) {
			return false
		}
	}
	return true
}

func validateToken(token string) bool {
	result, _ := regexp.MatchString(`(?i)^(RAW_TEXT|RECORD|(RECORD\..+?)) (CONTAINS|NOT_CONTAINS|EQUALS|NOT_EQUALS) (DATE|TIME|CURRENCY|NUMBER|FHS|FEDS|REGEX|IPV4|IPV6|)((\(.+?\))|(["']{1}.+?["']{1})|([\?\w\d]+))$`, token)
	return result
}

func validateTree(node *Node) bool {
	if isOperator(node.Expression) && !node.hasSubnodes() {
		return false
	}
	return true
}
