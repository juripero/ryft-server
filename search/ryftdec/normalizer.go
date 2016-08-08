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
	_ "fmt"
	"strings"
)

func normalizeTree(node *Node, booleansLimit map[string]int) {
	sameLevelNormalization(node, booleansLimit)
	if node.hasSubnodes() {
		differentLevelNormalization(node, booleansLimit)
	}
}

func sameLevelNormalization(node *Node, booleansLimit map[string]int) {
	if node.isNormalizable(booleansLimit) {
		subnodesType := node.SubNodes[0].Type
		node.Expression = node.SubNodes[0].Expression + " " + node.Expression + " " + node.SubNodes[1].Expression
		node.Type = subnodesType
		node.SubNodes = node.SubNodes[0:0]

		// Parent node changed, try to normalize it too
		if (node.Parent != nil) && node.Parent.hasSubnodes() {
			normalizeTree(node.Parent, booleansLimit)
		}
	} else {
		for _, subNode := range node.SubNodes {
			normalizeTree(subNode, booleansLimit)
		}
	}
}

func differentLevelNormalization(node *Node, booleansLimit map[string]int) {
	leftSubnode := node.SubNodes[0]
	rightSubnode := node.SubNodes[1]

	searchAndOperatorSubnodes := (leftSubnode.isSearch() && rightSubnode.isOperator()) || (leftSubnode.isOperator() && rightSubnode.isSearch())
	sameTypeOperators := (node.Type == leftSubnode.Type) || (node.Type == rightSubnode.Type)
	sameTypeQueries := queriesWithSameType(leftSubnode, rightSubnode) || queriesWithSameType(rightSubnode, leftSubnode)
	subnodeIsNormalizable := sameTypeOperators && sameTypeQueries && node.boolLimitIsNotReached(booleansLimit)

	if node.hasSubnodes() && searchAndOperatorSubnodes && subnodeIsNormalizable {
		if leftSubnode.isSearch() {
			appendNode(rightSubnode, leftSubnode)
		}
	}
}

func queriesWithSameType(node1 *Node, node2 *Node) bool {
	for _, node := range node2.SubNodes {
		return (node1.Type == node.Type) && node1.optionsEqual(node)
	}
	return false
}

func appendNode(srcParentNode, dstNode *Node) {
	srcNode, otherNode := splitNodes(dstNode, srcParentNode)

	dstNode.Expression = dstNode.Expression + " " + dstNode.Parent.Expression + " " + srcNode.Expression
	srcParentNode.Expression = otherNode.Expression
	srcParentNode.Type = otherNode.Type
	srcParentNode.SubNodes = srcParentNode.SubNodes[0:0]
}

func splitNodes(dstNode, parentNode *Node) (*Node, *Node) {
	var srcNode, otherNode *Node
	for _, node := range parentNode.SubNodes {
		if node.Type == dstNode.Type {
			srcNode = node
		} else {
			otherNode = node
		}
	}
	return srcNode, otherNode
}

func countBoolOperators(node *Node) int {
	return strings.Count(node.Expression, " OR ") + strings.Count(node.Expression, " AND ")
}
