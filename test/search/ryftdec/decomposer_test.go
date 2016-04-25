//Tests for follwing tree
//Expression: 'AND'
//  Expression: 'AND'
//    Expression: '(RECORD.id CONTAINS TIME("1003"))'
//    Expression: '(RECORD.id CONTAINS DATE("100301"))'
//  Expression: 'AND'
//    Expression: '(RECORD.id CONTAINS TIME("200"))'
//    Expression: 'AND'
//      Expression: '(RECORD.id CONTAINS DATE("300"))'
//      Expression: '(RECORD.id CONTAINS DATE("400"))'

package ryftdec

import (
	"testing"

	"github.com/getryft/ryft-server/search/ryftdec"
)

func TestNode1(t *testing.T) {
	result := tree()
	node := result
	if len(node.SubNodes) != 2 {
		t.Error("Expected 2 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != "AND" {
		t.Error("Expected AND, got", node.Expression)
	}
}

func TestNode2(t *testing.T) {
	result := tree()
	node := result.SubNodes[0]

	if len(node.SubNodes) != 2 {
		t.Error("Expected 2 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != "AND" {
		t.Error("Expected AND, got", node.Expression)
	}
}

func TestNode3(t *testing.T) {
	result := tree()
	node := result.SubNodes[1]

	if len(node.SubNodes) != 2 {
		t.Error("Expected 2 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != "AND" {
		t.Error("Expected AND, got", node.Expression)
	}
}

func TestNode4(t *testing.T) {
	result := tree()
	node := result.SubNodes[0].SubNodes[0]

	if len(node.SubNodes) != 0 {
		t.Error("Expected 0 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != `(RECORD.id CONTAINS TIME("1003"))` {
		t.Error(`Expected (RECORD.id CONTAINS TIME("1003")), got`, node.Expression)
	}
}

func TestNode5(t *testing.T) {
	result := tree()
	node := result.SubNodes[0].SubNodes[1]

	if len(node.SubNodes) != 0 {
		t.Error("Expected 0 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != `(RECORD.id CONTAINS DATE("100301"))` {
		t.Error(`Expected (RECORD.id CONTAINS DATE("100301")), got`, node.Expression)
	}
}

func TestNode6(t *testing.T) {
	result := tree()
	node := result.SubNodes[1].SubNodes[0]

	if len(node.SubNodes) != 0 {
		t.Error("Expected 0 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != `(RECORD.id CONTAINS TIME("200"))` {
		t.Error(`Expected (RECORD.id CONTAINS TIME("200")), got`, node.Expression)
	}
}

func TestNode7(t *testing.T) {
	result := tree()
	node := result.SubNodes[1].SubNodes[1]

	if len(node.SubNodes) != 2 {
		t.Error("Expected 2 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != `AND` {
		t.Error(`Expected AND, got`, node.Expression)
	}
}

func TestNode8(t *testing.T) {
	result := tree()
	node := result.SubNodes[1].SubNodes[1].SubNodes[0]

	if len(node.SubNodes) != 0 {
		t.Error("Expected 0 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != `(RECORD.id CONTAINS DATE("300"))` {
		t.Error(`Expected (RECORD.id CONTAINS DATE("300")), got`, node.Expression)
	}
}

func TestNode9(t *testing.T) {
	result := tree()
	node := result.SubNodes[1].SubNodes[1].SubNodes[1]

	if len(node.SubNodes) != 0 {
		t.Error("Expected 0 subnodes, got", len(node.SubNodes))
	}
	if node.Expression != `(RECORD.id CONTAINS DATE("400"))` {
		t.Error(`Expected (RECORD.id CONTAINS DATE("400")), got`, node.Expression)
	}
}

func tree() *ryftdec.Node {
	query := `((RECORD.id CONTAINS TIME("1003"))AND(RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`
	result, _ := ryftdec.Decompose(query)
	return result
}
